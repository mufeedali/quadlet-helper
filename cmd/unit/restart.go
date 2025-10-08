package unit

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:               "restart <unit-name>",
	Short:             "Restart a quadlet unit",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		unitName := args[0]
		serviceName := fmt.Sprintf("%s.service", unitName)

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Restarting %s...", serviceName)))

		c := exec.Command("systemctl", "--user", "restart", serviceName)
		output, err := c.CombinedOutput()
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error restarting service: %v\n%s", err, string(output))))
			os.Exit(1)
		}

		fmt.Println(string(output))
		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully restarted %s", serviceName)))
	},
}
