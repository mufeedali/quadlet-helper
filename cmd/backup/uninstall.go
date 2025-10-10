package backup

import (
	"fmt"
	"os"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:               "uninstall [backup-name]",
	Short:             "Uninstall backup service and timer from systemd",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getInstalledBackupCompletions(),
	Run: func(cmd *cobra.Command, args []string) {
		backupName := args[0]

		// Check if actually installed
		timerPath, _ := backup.GetTimerFilePath(backupName)
		if !fileExists(timerPath) {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Backup '%s' is not installed", backupName)))
			fmt.Println("\nTo install it, use:")
			fmt.Printf("  qh backup install %s\n", backupName)
			os.Exit(1)
		}

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Uninstalling backup: %s", backupName)))

		// Stop and disable timer
		runSystemctl("stop", fmt.Sprintf("%s-backup.timer", backupName))
		runSystemctl("disable", fmt.Sprintf("%s-backup.timer", backupName))

		// Stop service if running
		runSystemctl("stop", fmt.Sprintf("%s-backup.service", backupName))

		// Remove service file
		serviceFilePath, _ := backup.GetServiceFilePath(backupName)
		if err := os.Remove(serviceFilePath); err != nil {
			if !os.IsNotExist(err) {
				fmt.Println(shared.WarningStyle.Render(fmt.Sprintf("Warning: could not remove %s: %v", serviceFilePath, err)))
			}
		} else {
			fmt.Println(shared.CheckMark + " Removed " + shared.FilePathStyle.Render(serviceFilePath))
		}

		// Remove timer file
		timerFilePath, _ := backup.GetTimerFilePath(backupName)
		if err := os.Remove(timerFilePath); err != nil {
			if !os.IsNotExist(err) {
				fmt.Println(shared.WarningStyle.Render(fmt.Sprintf("Warning: could not remove %s: %v", timerFilePath, err)))
			}
		} else {
			fmt.Println(shared.CheckMark + " Removed " + shared.FilePathStyle.Render(timerFilePath))
		}

		// Remove notification service if exists
		notifyFilePath, _ := backup.GetNotificationServiceFilePath(backupName)
		if err := os.Remove(notifyFilePath); err == nil {
			fmt.Println(shared.CheckMark + " Removed " + shared.FilePathStyle.Render(notifyFilePath))
		}

		// Reload systemd
		runSystemctl("daemon-reload")

		fmt.Println(shared.SuccessStyle.Render("\n✓ Uninstallation complete!"))
		fmt.Println("\nNote: Configuration file still exists. To remove it, run:")
		configPath, _ := backup.GetConfigPath(backupName)
		fmt.Printf("  rm %s\n", configPath)
	},
}
