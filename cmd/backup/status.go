package backup

import (
	"fmt"

	internalbackup "github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:               "status [backup-name]",
	Short:             "Show backup service and timer status",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupName := args[0]

		if _, err := loadBackupConfig(backupName); err != nil {
			return err
		}

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Status for backup: %s", backupName)))
		fmt.Println()

		if !isInstalledBackup(backupName) {
			fmt.Println(shared.WarningStyle.Render("Backup is not installed"))
			fmt.Printf("\nTo install, run: qh backup install %s\n", backupName)
			return nil
		}

		timerName := internalbackup.BackupTimerName(backupName)
		serviceName := internalbackup.BackupServiceName(backupName)

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
		listOutput, err := systemd.ListTimers(timerName)
		fmt.Print(listOutput)
		if err != nil {
			return cmdutil.Wrap(err, "getting timer schedule")
		}
		return nil
	},
}
