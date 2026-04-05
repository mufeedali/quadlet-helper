package backup

import (
	"fmt"

	internalbackup "github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:               "run [backup-name]",
	Short:             "Run a backup immediately",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupName := args[0]

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Running backup: %s", backupName)))
		fmt.Println()

		config, err := loadBackupConfig(backupName)
		if err != nil {
			return err
		}

		result, err := internalbackup.Run(config, false)
		if err != nil {
			if config.Notifications.Enabled && config.Notifications.OnFailure {
				_ = internalbackup.SendNotification(config, "failure", fmt.Sprintf("Error: %v\n\nOutput:\n%s", err, result.Output))
			}
			return cmdutil.Wrap(err, "backup failed")
		}

		fmt.Println()
		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Backup completed successfully in %.2f seconds", result.EndTime.Sub(result.StartTime).Seconds())))

		if config.Notifications.Enabled && config.Notifications.OnSuccess {
			_ = internalbackup.SendNotification(config, "success", result.Output)
		}

		return nil
	},
}
