package clusters_test

import (
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/Kong/kuma/pkg/xds/envoy"
	"github.com/Kong/kuma/pkg/xds/envoy/clusters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("LbSubset", func() {

	type testCase struct {
		clusterName string
		tags        []envoy.Tags
		expected    string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			cluster, err := clusters.NewClusterBuilder().
				Configure(clusters.EdsCluster(given.clusterName)).
				Configure(clusters.LbSubset(given.tags)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("LbSubset is empty if there is only a service tag", testCase{
			clusterName: "backend",
			tags: []envoy.Tags{
				{
					"service": "backend",
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
            name: backend
            type: EDS`,
		}),
		Entry("LbSubset is set when more than service tag is set", testCase{
			clusterName: "backend",
			tags: []envoy.Tags{
				{
					"service": "backend",
					"version": "v1",
				},
				{
					"service": "backend",
					"version": "v2",
				},
				{
					"service": "backend",
					"version": "v3",
					"cluster": "k8s-1",
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
            lbSubsetConfig:
              fallbackPolicy: ANY_ENDPOINT
              subsetSelectors:
              - fallbackPolicy: NO_FALLBACK
                keys:
                - version
              - fallbackPolicy: NO_FALLBACK
                keys:
                - cluster
                - version
            name: backend
            type: EDS`,
		}),
	)
})