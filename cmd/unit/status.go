package unit

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:               "status <unit-name>",
	Short:             "Get the status of a quadlet unit",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		unitName := args[0]
		serviceName := fmt.Sprintf("%s.service", unitName)

		c := exec.Command("systemctl", "--user", "status", serviceName)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		// Status command returns non-zero exit code when service is not running, so we don't exit on error
		_ = c.Run()
	},
}
