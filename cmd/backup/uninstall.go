package backup

import (
	"fmt"

	internalbackup "github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:               "uninstall [backup-name]",
	Short:             "Uninstall backup service and timer from systemd",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getInstalledBackupCompletions(),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupName := args[0]

		if !isInstalledBackup(backupName) {
			return cmdutil.Errorf("backup %q is not installed\n\nTo install it, use:\n  qh backup install %s", backupName, backupName)
		}

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Uninstalling backup: %s", backupName)))

		result, err := systemd.UninstallUserUnits(
			[]string{internalbackup.BackupTimerName(backupName), internalbackup.BackupServiceName(backupName)},
			[]string{internalbackup.BackupTimerName(backupName)},
			[]string{
				internalbackup.BackupServiceName(backupName),
				internalbackup.BackupTimerName(backupName),
				fmt.Sprintf("backup-notify@%s.service", backupName),
			},
		)
		if err != nil {
			return err
		}
		for _, warning := range result.Warnings {
			fmt.Println(shared.WarningStyle.Render("Warning: " + warning.Error()))
		}
		for _, path := range result.RemovedPaths {
			fmt.Println(shared.CheckMark + " Removed " + shared.FilePathStyle.Render(path))
		}

		fmt.Println(shared.SuccessStyle.Render("\n✓ Uninstallation complete!"))
		fmt.Println("\nNote: Configuration file still exists. To remove it, run:")
		configPath, _ := internalbackup.GetConfigPath(backupName)
		fmt.Printf("  rm %s\n", configPath)
		return nil
	},
}
