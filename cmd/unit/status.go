package unit

import (
	"fmt"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:               "status <unit-name>...",
	Short:             "Get the status of one or more quadlet units",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		// Build service names from provided unit names
		services := make([]string, len(args))
		for i, unitName := range args {
			services[i] = fmt.Sprintf("%s.service", unitName)
		}

		// Run systemctl status for all services in one call
		output, _ := systemd.Status(services...)
		fmt.Println(output)

		// Print a short summary line for convenience
		fmt.Println(strings.Join(services, ", "))
	},
}
