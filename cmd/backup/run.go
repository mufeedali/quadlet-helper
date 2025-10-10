package backup

import (
	"fmt"
	"os"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:               "run [backup-name]",
	Short:             "Run a backup immediately",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	Run: func(cmd *cobra.Command, args []string) {
		backupName := args[0]

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Running backup: %s", backupName)))
		fmt.Println()

		// Load config
		config, err := backup.LoadConfig(backupName)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error loading config: %v", err)))
			os.Exit(1)
		}

		// Run backup
		result, err := backup.Run(config, false)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Backup failed: %v", err)))
			fmt.Println("\nOutput:")
			fmt.Println(result.Output)

			// Send failure notification
			if config.Notifications.Enabled && config.Notifications.OnFailure {
				_ = backup.SendNotification(config, "failure", fmt.Sprintf("Error: %v\n\nOutput:\n%s", err, result.Output))
			}

			os.Exit(1)
		}

		fmt.Println(result.Output)
		fmt.Println()
		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Backup completed successfully in %.2f seconds", result.EndTime.Sub(result.StartTime).Seconds())))

		// Send success notification if enabled
		if config.Notifications.Enabled && config.Notifications.OnSuccess {
			_ = backup.SendNotification(config, "success", result.Output)
		}
	},
}
