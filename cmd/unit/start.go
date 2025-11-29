package unit

import (
	"fmt"
	"os"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:               "start <unit-name>...",
	Short:             "Start one or more quadlet units",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		// Resolve service names from provided unit names
		services, err := resolveServiceNames(args)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error resolving service names: %v", err)))
			os.Exit(1)
		}

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Starting %s...", strings.Join(services, " "))))

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

		// Build systemctl args and start all services in one call
		output, err := systemd.StartMultiple(services)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error starting services: %v\n%s", err, string(output))))
			os.Exit(1)
		}

		if len(output) > 0 {
			fmt.Println(string(output))
		}
		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully started %s", strings.Join(services, " "))))
	},
}

func init() {
	startCmd.Flags().BoolVar(&noReload, "no-reload", false, "Skip systemctl daemon-reload step")
}
