package unit

import (
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var stopTypes []string

var stopCmd = &cobra.Command{
	Use:               "stop <unit-name>...",
	Short:             "Stop one or more quadlet units",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: activeUnitCompletionFunc,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServiceAction(args, stopTypes, "Stopping %s...", false, false, systemd.StopMultiple, "stopping services", "✓ Successfully stopped %s", nil)
	},
}

func init() {
	stopCmd.Flags().StringSliceVar(&stopTypes, "type", []string{"container", "kube", "pod"}, "Quadlet unit types to act on")
	_ = stopCmd.RegisterFlagCompletionFunc("type", typeCompletionFunc)
}
