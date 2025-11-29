package unit

import (
	"fmt"
	"os"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:               "stop <unit-name>...",
	Short:             "Stop one or more quadlet units",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: activeUnitCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		// Resolve service names from provided unit names
		services, err := resolveServiceNames(args)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error resolving service names: %v", err)))
			os.Exit(1)
		}

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Stopping %s...", strings.Join(services, " "))))

		// Build systemctl args and stop all services in one call
		output, err := systemd.StopMultiple(services)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error stopping services: %v\n%s", err, string(output))))
			os.Exit(1)
		}

		if len(output) > 0 {
			fmt.Println(string(output))
		}
		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully stopped %s", strings.Join(services, ", "))))
	},
}
