package unit

import (
	"fmt"
	"os"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var noReload bool

var restartCmd = &cobra.Command{
	Use:               "restart <unit-name>...",
	Short:             "Restart one or more quadlet units",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		// Resolve service names from provided unit names
		services, err := resolveServiceNames(args)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error resolving service names: %v", err)))
			os.Exit(1)
		}

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Restarting %s...", strings.Join(services, " "))))

		// Reload systemctl daemon (unless --no-reload)
		if !noReload {
			fmt.Println(shared.TitleStyle.Render("Reloading systemctl daemon..."))
			output, err := systemd.DaemonReload()
			if err != nil {
				fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error reloading systemctl daemon: %v\n%s", err, string(output))))
				fmt.Println(shared.InfoMark + " Please reload manually: " + "systemctl --user daemon-reload")
				os.Exit(1)
			}
			if len(output) > 0 {
				fmt.Println(string(output))
			}
			fmt.Println(shared.SuccessStyle.Render("✓ systemctl daemon reloaded"))
		} else {
			fmt.Println(shared.InfoMark + " Skipping systemctl daemon reload ( --no-reload )")
		}

		// Restart services
		output, err := systemd.RestartMultiple(services)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error restarting services: %v\n%s", err, string(output))))
			fmt.Println(shared.InfoMark + " Please restart manually: " + "systemctl --user restart " + strings.Join(services, " "))
			os.Exit(1)
		}
		if len(output) > 0 {
			fmt.Println(string(output))
		}
		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully restarted %s", strings.Join(services, ", "))))
	},
}

func init() {
	restartCmd.Flags().BoolVar(&noReload, "no-reload", false, "Skip systemctl daemon-reload step")
}
