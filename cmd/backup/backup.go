package backup

import (
	"github.com/spf13/cobra"
)

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage custom systemd backup services",
	Long:  `This command helps manage custom systemd user services and timers for backups using rsync, restic, or rclone.`,
}

func init() {
	BackupCmd.AddCommand(createCmd)
	BackupCmd.AddCommand(installCmd)
	BackupCmd.AddCommand(uninstallCmd)
	BackupCmd.AddCommand(listCmd)
	BackupCmd.AddCommand(runCmd)
	BackupCmd.AddCommand(verifyCmd)
	BackupCmd.AddCommand(testCmd)
	BackupCmd.AddCommand(statusCmd)
	BackupCmd.AddCommand(logsCmd)
	BackupCmd.AddCommand(editCmd)
	BackupCmd.AddCommand(notifyCmd)
	BackupCmd.AddCommand(cleanupCmd)
}
