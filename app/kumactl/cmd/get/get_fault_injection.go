package get

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"

	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

func newGetFaultInjectionCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fault-injection NAME",
		Short: "Show a single Fault-Injection resource",
		Long:  `Show a single Fault-Injection resource.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			name := args[0]
			currentMesh := pctx.CurrentMesh()
			faultInjection := &mesh.FaultInjectionResource{}
			if err := rs.Get(context.Background(), faultInjection, store.GetByKey(name, currentMesh)); err != nil {
				if store.IsResourceNotFound(err) {
					return errors.Errorf("No resources found in %s mesh", currentMesh)
				}
				return errors.Wrapf(err, "failed to get mesh %s", currentMesh)
			}
			faultInjections := &mesh.FaultInjectionResourceList{
				Items: []*mesh.FaultInjectionResource{faultInjection},
			}
			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printFaultInjections(pctx.Now(), faultInjections, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(faultInjection), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
