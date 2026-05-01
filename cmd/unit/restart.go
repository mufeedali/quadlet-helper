package unit

import (
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var restartNoReload bool
var restartTypes []string

var restartCmd = &cobra.Command{
	Use:               "restart <unit-name>...",
	Short:             "Restart one or more quadlet units",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServiceAction(
			args,
			restartTypes,
			"Restarting %s...",
			restartNoReload,
			true,
			systemd.RestartMultiple,
			"restarting services",
			"✓ Successfully restarted %s",
			func(services []string) string {
				return "Please restart manually: systemctl --user restart " + strings.Join(services, " ")
			},
		)
	},
}

func init() {
	restartCmd.Flags().BoolVar(&restartNoReload, "no-reload", false, "Skip systemctl daemon-reload step")
	restartCmd.Flags().StringSliceVar(&restartTypes, "type", []string{"container", "kube", "pod"}, "Quadlet unit types to act on")
	_ = restartCmd.RegisterFlagCompletionFunc("type", typeCompletionFunc)
}
