package unit

import (
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var startNoReload bool
var startTypes []string

var startCmd = &cobra.Command{
	Use:               "start <unit-name>...",
	Short:             "Start one or more quadlet units",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServiceAction(args, startTypes, "Starting %s...", startNoReload, true, systemd.StartMultiple, "starting services", "✓ Successfully started %s", nil)
	},
}

func init() {
	startCmd.Flags().BoolVar(&startNoReload, "no-reload", false, "Skip systemctl daemon-reload step")
	startCmd.Flags().StringSliceVar(&startTypes, "type", []string{"container", "kube", "pod"}, "Quadlet unit types to act on")
	_ = startCmd.RegisterFlagCompletionFunc("type", typeCompletionFunc)
}
