package backup

import (
	"fmt"

	internalbackup "github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:               "notify [backup-name] [status]",
	Short:             "Send email notification (used by systemd)",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: getNotifyCompletions(),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupName := args[0]
		status := args[1] // "success" or "failure"

		config, err := loadBackupConfig(backupName)
		if err != nil {
			return err
		}

		logs, err := internalbackup.GetLastBackupLog(backupName)
		if err != nil {
			logs = fmt.Sprintf("Failed to retrieve logs: %v", err)
		}

		if err := internalbackup.SendNotification(config, status, logs); err != nil {
			return cmdutil.Wrap(err, "sending notification")
		}

		fmt.Println(shared.SuccessStyle.Render("✓ Notification sent"))
		return nil
	},
}
