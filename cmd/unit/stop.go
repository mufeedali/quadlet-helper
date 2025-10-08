package unit

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:               "stop <unit-name>",
	Short:             "Stop a quadlet unit",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		unitName := args[0]
		serviceName := fmt.Sprintf("%s.service", unitName)

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Stopping %s...", serviceName)))

		c := exec.Command("systemctl", "--user", "stop", serviceName)
		output, err := c.CombinedOutput()
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error stopping service: %v\n%s", err, string(output))))
			os.Exit(1)
		}

		fmt.Println(string(output))
		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully stopped %s", serviceName)))
	},
}
