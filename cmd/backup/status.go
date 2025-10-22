package backup

import (
	"fmt"
	"os"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:               "status [backup-name]",
	Short:             "Show backup service and timer status",
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

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Status for backup: %s", backupName)))
		fmt.Println()

		// Check if timer exists
		timerFilePath, _ := backup.GetTimerFilePath(backupName)
		if _, err := os.Stat(timerFilePath); os.IsNotExist(err) {
			fmt.Println(shared.WarningStyle.Render("Backup is not installed"))
			fmt.Printf("\nTo install, run: qh backup install %s\n", backupName)
			return
		}

		// Show timer status
		timerName := backup.BackupTimerName(backupName)
		serviceName := backup.BackupServiceName(backupName)

		fmt.Println(shared.TitleStyle.Render("Timer:"))
		output, _ := systemd.Status(timerName)
		fmt.Println(output)

		fmt.Println()
		fmt.Println(shared.TitleStyle.Render("Service:"))
		output, _ = systemd.Status(serviceName)
		fmt.Println(output)

		// Show next run time
		fmt.Println()
		fmt.Println(shared.TitleStyle.Render("Schedule:"))
		listOutput, _ := systemd.ListTimers(timerName)
		fmt.Print(listOutput)
	},
}
