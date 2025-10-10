package backup

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
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
		fmt.Println(shared.TitleStyle.Render("Timer:"))
		runSystemctl("--no-pager", "status", fmt.Sprintf("%s-backup.timer", backupName))

		fmt.Println()
		fmt.Println(shared.TitleStyle.Render("Service:"))
		runSystemctl("--no-pager", "status", fmt.Sprintf("%s-backup.service", backupName))

		// Show next run time
		fmt.Println()
		fmt.Println(shared.TitleStyle.Render("Schedule:"))
		listCmd := exec.Command("systemctl", "--user", "list-timers", fmt.Sprintf("%s-backup.timer", backupName), "--no-pager")
		output, _ := listCmd.CombinedOutput()
		fmt.Print(string(output))
	},
}
