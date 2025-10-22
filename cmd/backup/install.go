package backup

import (
	"fmt"
	"os"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:               "install [backup-name]",
	Short:             "Install backup service and timer to systemd",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getNotInstalledBackupCompletions(),
	Run: func(cmd *cobra.Command, args []string) {
		backupName := args[0]

		// Check if already installed
		timerPath, _ := backup.GetTimerFilePath(backupName)
		if shared.FileExists(timerPath) {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Backup '%s' is already installed", backupName)))
			fmt.Println("\nTo reinstall, first uninstall it with:")
			fmt.Printf("  qh backup uninstall %s\n", backupName)
			os.Exit(1)
		}

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Installing backup: %s", backupName)))

		// Load config
		config, err := backup.LoadConfig(backupName)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error loading config: %v", err)))
			os.Exit(1)
		}

		// Get systemd user directory
		systemdDir, err := backup.GetSystemdUserDir()
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error: %v", err)))
			os.Exit(1)
		}

		if err := os.MkdirAll(systemdDir, 0755); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error creating systemd directory: %v", err)))
			os.Exit(1)
		}

		// Get executable path
		executablePath, err := os.Executable()
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error finding executable: %v", err)))
			os.Exit(1)
		}

		configPath, _ := backup.GetConfigPath(backupName)

		// Create service file
		serviceContent := backup.GetServiceTemplate(executablePath, configPath, backupName, config)
		serviceFilePath, _ := backup.GetServiceFilePath(backupName)
		if err := os.WriteFile(serviceFilePath, []byte(serviceContent), 0644); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error writing service file: %v", err)))
			os.Exit(1)
		}
		fmt.Println(shared.CheckMark + " Created " + shared.FilePathStyle.Render(serviceFilePath))

		// Create timer file
		timerContent, err := backup.GetTimerTemplate(backupName, config.Schedule)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error creating timer template: %v", err)))
			os.Exit(1)
		}
		timerFilePath, _ := backup.GetTimerFilePath(backupName)
		if err := os.WriteFile(timerFilePath, []byte(timerContent), 0644); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error writing timer file: %v", err)))
			os.Exit(1)
		}
		fmt.Println(shared.CheckMark + " Created " + shared.FilePathStyle.Render(timerFilePath))

		// Reload systemd
		if _, err := systemd.DaemonReload(); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error reloading systemd: %v", err)))
			os.Exit(1)
		}

		// Enable and start timer
		timerName := backup.BackupTimerName(backupName)
		if _, err := systemd.Enable(timerName); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error enabling timer: %v", err)))
			os.Exit(1)
		}
		if _, err := systemd.Start(timerName); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error starting timer: %v", err)))
			os.Exit(1)
		}

		fmt.Println(shared.SuccessStyle.Render("\n✓ Installation complete!"))
		fmt.Println(shared.TitleStyle.Render("Timer status:"))
		output, _ := systemd.Status(timerName)
		fmt.Println(output)
	},
}
