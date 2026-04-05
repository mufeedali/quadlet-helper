package backup

import (
	"fmt"

	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:               "logs [backup-name]",
	Short:             "View backup logs",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupName := args[0]

		if _, err := loadBackupConfig(backupName); err != nil {
			return err
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

		if err := runJournalctl(journalArgs); err != nil {
			return cmdutil.Wrap(err, "viewing logs")
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().IntP("lines", "n", 50, "Number of log lines to show")
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
}
