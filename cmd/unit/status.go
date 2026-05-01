package unit

import (
	"fmt"

	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var statusTypes []string

var statusCmd = &cobra.Command{
	Use:               "status <unit-name>...",
	Short:             "Get the status of one or more quadlet units",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	RunE: func(cmd *cobra.Command, args []string) error {
		services, err := loadServices(args, statusTypes)
		if err != nil {
			return err
		}

		output, _ := systemd.Status(services...)
		fmt.Println(output)
		if output == "" {
			return cmdutil.Errorf("no status output returned")
		}
		if _, err := systemd.IsActiveMultiple(services); err != nil {
			return cmdutil.Wrap(err, "getting unit status")
		}
		return nil
	},
}

func init() {
	statusCmd.Flags().StringSliceVar(&statusTypes, "type", []string{"container", "kube", "pod"}, "Quadlet unit types to act on")
	_ = statusCmd.RegisterFlagCompletionFunc("type", typeCompletionFunc)
}
