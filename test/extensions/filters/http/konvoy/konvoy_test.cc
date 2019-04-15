#include "extensions/filters/http/konvoy/konvoy.h"

#include "common/protobuf/utility.h"

#include "test/mocks/grpc/mocks.h"
#include "test/mocks/http/mocks.h"
#include "test/mocks/stats/mocks.h"
#include "test/test_common/simulated_time_system.h"
#include "test/test_common/utility.h"

#include "gmock/gmock.h"
#include "gtest/gtest.h"

using KonvoyProtoConfig = envoy::config::filter::http::konvoy::v2alpha::Konvoy;
using envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage;
using envoy::service::konvoy::v2alpha::ProxyHttpRequestServerMessage;
using testing::NiceMock;
using testing::ReturnRef;
using testing::StrictMock;
using testing::WhenDynamicCastTo;

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

class KonvoyHttpFilterTest : public testing::Test {
public:
    ConfigSharedPtr setupConfig(std::function<void(KonvoyProtoConfig&)> configure) {
      KonvoyProtoConfig config;
      config.set_stat_prefix("demo-grpc-server.");
      if (configure) {
        configure(config);
      }
      return std::make_shared<Config>(config, store_, time_system_);
    }

    KonvoyHttpFilterTest() : config_(setupConfig(nullptr)), async_client_(new StrictMock<Grpc::MockAsyncClient>()) {
      filter_ = std::make_unique<Filter>(config_, Grpc::AsyncClientPtr{async_client_});
      filter_->setDecoderFilterCallbacks(callbacks_);
    }

    Stats::IsolatedStoreImpl store_;
    Event::SimulatedTimeSystem time_system_;
    ConfigSharedPtr config_;
    StrictMock<Grpc::MockAsyncClient>* async_client_;
    StrictMock<Grpc::MockAsyncStream> async_stream_;

    StrictMock<Http::MockStreamDecoderFilterCallbacks> callbacks_;
    Buffer::OwnedImpl data_;

    std::unique_ptr<Filter> filter_;
    Http::TestHeaderMapImpl headers_;
};

TEST_F(KonvoyHttpFilterTest, KonvoyServiceIsDown) {
  // expect : an attempt to open a new stream to Http Konvoy Service to fail
  EXPECT_CALL(*async_client_, start(_, _))
          .WillOnce(Invoke([this](const Protobuf::MethodDescriptor&, Grpc::AsyncStreamCallbacks&) {
              // simulate unavailable cluster
              filter_->onRemoteClose(Grpc::Status::GrpcStatus::Unavailable, "Cluster not available");
              return nullptr;
          }));
  // and expect : response with HTTP 500
  Http::TestHeaderMapImpl response_headers{{":status", "500"}};
  EXPECT_CALL(callbacks_, encodeHeaders_(HeaderMapEqualRef(&response_headers), true));
  // when : HTTP request headers received
  EXPECT_EQ(Http::FilterHeadersStatus::StopIteration, filter_->decodeHeaders(headers_, true));
  // then : metrics get updated
  EXPECT_EQ(0U, config_->stats().rq_active_.value());
  EXPECT_EQ(1U, config_->stats().rq_total_.value());
  EXPECT_EQ(1U, config_->stats().rq_error_.value());
  EXPECT_EQ(0U, config_->stats().rq_cancel_.value());
}

TEST_F(KonvoyHttpFilterTest, MutateGetRequest) {
  // given : simple GET request
  headers_.addCopy(":authority", "example.org");
  // expect : an attempt to open a new stream to Http Konvoy Service to succeed
  EXPECT_CALL(*async_client_, start(_, _))
          .WillOnce(Invoke([this](const Protobuf::MethodDescriptor&, Grpc::AsyncStreamCallbacks&) {
              return &this->async_stream_;
          }));
  // and expect : request headers to be forwarded to Http Konvoy Service
  ProxyHttpRequestClientMessage client_request_headers_message{};
  auto* original_authority_header = client_request_headers_message.mutable_request_headers()->mutable_headers()->add_headers();
  original_authority_header->set_key(":authority");
  original_authority_header->set_value("example.org");
  EXPECT_CALL(async_stream_, sendMessage(WhenDynamicCastTo<const ProxyHttpRequestClientMessage&>(ProtoEq(client_request_headers_message)), false));
  // and expect : pseudo request trailers to be forwarded to Http Konvoy Service
  ProxyHttpRequestClientMessage client_request_trailers_message{};
  client_request_trailers_message.mutable_request_trailers();
  EXPECT_CALL(async_stream_, sendMessage(WhenDynamicCastTo<const ProxyHttpRequestClientMessage&>(ProtoEq(client_request_trailers_message)), true));
  // when : HTTP request headers received
  EXPECT_EQ(Http::FilterHeadersStatus::StopIteration, filter_->decodeHeaders(headers_, true));
  // and then : metrics get updated
  EXPECT_EQ(1U, config_->stats().rq_active_.value());
  EXPECT_EQ(1U, config_->stats().rq_total_.value());
  EXPECT_EQ(0U, config_->stats().rq_error_.value());
  EXPECT_EQ(0U, config_->stats().rq_cancel_.value());

  // given : Http Konvoy Service mutates request headers
  auto server_message = std::make_unique<ProxyHttpRequestServerMessage>();
  auto* modified_authority_header = server_message->mutable_request_headers()->mutable_headers()->add_headers();;
  modified_authority_header->set_key(":authority");
  modified_authority_header->set_value("httpbin.org");
  // when : a message with mutated request headers is received back from Http Konvoy Service
  EXPECT_NO_THROW(filter_->onReceiveMessage(std::move(server_message)));
  // then : metrics don't change
  EXPECT_EQ(1U, config_->stats().rq_active_.value());
  EXPECT_EQ(1U, config_->stats().rq_total_.value());
  EXPECT_EQ(0U, config_->stats().rq_error_.value());
  EXPECT_EQ(0U, config_->stats().rq_cancel_.value());

  // expect : filter chain to proceed
  EXPECT_CALL(callbacks_, continueDecoding());
  // when : a stream is closed by Http Konvoy Service
  EXPECT_NO_THROW(filter_->onRemoteClose(Grpc::Status::GrpcStatus::Ok, "OK"));
  // then : metrics get updated
  EXPECT_EQ(0U, config_->stats().rq_active_.value());
  EXPECT_EQ(1U, config_->stats().rq_total_.value());
  EXPECT_EQ(0U, config_->stats().rq_error_.value());
  EXPECT_EQ(0U, config_->stats().rq_cancel_.value());
}

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
