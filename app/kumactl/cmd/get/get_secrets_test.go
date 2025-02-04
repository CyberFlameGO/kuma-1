package get_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("kumactl get secrets", func() {

	var sampleSecrets []*system.SecretResource

	BeforeEach(func() {
		sampleSecrets = []*system.SecretResource{
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "sec-1",
				},
				Spec: &system_proto.Secret{
					Data: util_proto.Bytes([]byte("test")),
				},
			},
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "sec-2",
				},
				Spec: &system_proto.Secret{
					Data: util_proto.Bytes([]byte("test2")),
				},
			},
		}
	})

	Describe("GetSecretsCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			store = core_store.NewPaginationStore(memory_resources.NewStore())

			rootCtx, err := test_kumactl.MakeRootContext(rootTime, store, system.SecretResourceTypeDescriptor)
			Expect(err).ToNot(HaveOccurred())

			for _, pt := range sampleSecrets {
				key := core_model.ResourceKey{
					Mesh: pt.Meta.GetMesh(),
					Name: pt.Meta.GetName(),
				}
				err := store.Create(context.Background(), pt, core_store.CreateBy(key))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
		})

		type testCase struct {
			outputFormat string
			goldenFile   string
			matcher      func(interface{}) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get secrets -o table|json|yaml",
			func(given testCase) {
				// when
				Expect(
					ExecuteRootCommand(rootCmd, "secrets", given.outputFormat, ""),
				).To(Succeed())

				// when
				expected, err := os.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(buf.String()).To(given.matcher(expected))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-secrets.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-secrets.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-secrets.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-secrets.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})
})
