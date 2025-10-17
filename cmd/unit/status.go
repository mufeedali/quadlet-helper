package unit

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
		cmdArgs := append([]string{"--user", "status"}, services...)
		c := exec.Command("systemctl", cmdArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		// Status command returns non-zero exit code when service is not running, so we don't exit on error
		_ = c.Run()

		// Print a short summary line for convenience
		fmt.Println(strings.Join(services, ", "))
	},
}
