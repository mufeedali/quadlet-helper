package backup

import (
	"fmt"
	"os"

	internalbackup "github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:               "install [backup-name]",
	Short:             "Install backup service and timer to systemd",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getNotInstalledBackupCompletions(),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupName := args[0]

		if isInstalledBackup(backupName) {
			return cmdutil.Errorf("backup %q is already installed\n\nTo reinstall, first uninstall it with:\n  qh backup uninstall %s", backupName, backupName)
		}

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Installing backup: %s", backupName)))

		config, err := loadBackupConfig(backupName)
		if err != nil {
			return err
		}

		executablePath, err := os.Executable()
		if err != nil {
			return cmdutil.Wrap(err, "finding executable")
		}

		timerContent, err := internalbackup.GetTimerTemplate(backupName, config.Schedule)
		if err != nil {
			return cmdutil.Wrap(err, "creating timer template")
		}

		paths, err := systemd.InstallUserUnits([]systemd.UserUnitFile{
			{Name: internalbackup.BackupServiceName(backupName), Content: internalbackup.GetServiceTemplate(executablePath, backupName, config), Mode: 0644},
			{Name: internalbackup.BackupTimerName(backupName), Content: timerContent, Mode: 0644},
		}, []string{internalbackup.BackupTimerName(backupName)})
		if err != nil {
			return err
		}
		for _, path := range paths {
			fmt.Println(shared.CheckMark + " Created " + shared.FilePathStyle.Render(path))
		}

		timerName := internalbackup.BackupTimerName(backupName)

		fmt.Println(shared.SuccessStyle.Render("\n✓ Installation complete!"))
		fmt.Println(shared.TitleStyle.Render("Timer status:"))
		output, err := systemd.Status(timerName)
		fmt.Println(output)
		if err != nil {
			return cmdutil.Wrap(err, "getting timer status")
		}
		active, err := systemd.IsActive(timerName)
		if err != nil {
			return cmdutil.Wrap(err, "checking timer active state")
		}
		if !active {
			return cmdutil.Errorf("timer %s did not become active after installation", timerName)
		}
		return nil
	},
}
