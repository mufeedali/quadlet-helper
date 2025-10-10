package backup

import (
	"fmt"
	"os"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:               "notify [backup-name] [status]",
	Short:             "Send email notification (used by systemd)",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: getNotifyCompletions(),
	Run: func(cmd *cobra.Command, args []string) {
		backupName := args[0]
		status := args[1] // "success" or "failure"

		// Load config
		config, err := backup.LoadConfig(backupName)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error loading config: %v", err)))
			os.Exit(1)
		}

		// Get logs
		logs, err := backup.GetLastBackupLog(backupName)
		if err != nil {
			logs = fmt.Sprintf("Failed to retrieve logs: %v", err)
		}

		// Send notification
		if err := backup.SendNotification(config, status, logs); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Failed to send notification: %v", err)))
			os.Exit(1)
		}

		fmt.Println(shared.SuccessStyle.Render("✓ Notification sent"))
	},
}
