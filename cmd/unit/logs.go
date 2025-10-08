package unit

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:               "logs <unit-name>",
	Short:             "View logs of a quadlet unit",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		unitName := args[0]
		serviceName := fmt.Sprintf("%s.service", unitName)
		follow, _ := cmd.Flags().GetBool("follow")

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Logs for %s...", serviceName)))

		args_list := []string{"--user", "-u", serviceName}
		if follow {
			args_list = append(args_list, "-f")
		}

		c := exec.Command("journalctl", args_list...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err := c.Run()
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error getting logs: %v", err)))
			os.Exit(1)
		}
	},
}

func init() {
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
}
