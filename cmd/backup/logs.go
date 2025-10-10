package backup

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:               "logs [backup-name]",
	Short:             "View backup logs",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	Run: func(cmd *cobra.Command, args []string) {
		backupName := args[0]

		// Load config to verify it exists
		_, err := backup.LoadConfig(backupName)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error loading config: %v", err)))
			os.Exit(1)
		}

		lines, _ := cmd.Flags().GetInt("lines")
		follow, _ := cmd.Flags().GetBool("follow")

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Logs for backup: %s", backupName)))
		fmt.Println()

		// Build journalctl command
		journalArgs := []string{
			"--user",
			"-u", fmt.Sprintf("%s-backup.service", backupName),
			"--no-pager",
		}

		if follow {
			journalArgs = append(journalArgs, "-f")
		} else {
			journalArgs = append(journalArgs, "-n", fmt.Sprintf("%d", lines))
		}

		// Run journalctl
		journalCmd := exec.Command("journalctl", journalArgs...)
		journalCmd.Stdout = os.Stdout
		journalCmd.Stderr = os.Stderr
		journalCmd.Stdin = os.Stdin

		if err := journalCmd.Run(); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error viewing logs: %v", err)))
			os.Exit(1)
		}
	},
}

func init() {
	logsCmd.Flags().IntP("lines", "n", 50, "Number of log lines to show")
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
}
