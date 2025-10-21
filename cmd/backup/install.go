package backup

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
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
		if fileExists(timerPath) {
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
		runSystemctl("daemon-reload")

		// Enable and start timer
		runSystemctl("enable", fmt.Sprintf("%s-backup.timer", backupName))
		runSystemctl("start", fmt.Sprintf("%s-backup.timer", backupName))

		fmt.Println(shared.SuccessStyle.Render("\n✓ Installation complete!"))
		fmt.Println(shared.TitleStyle.Render("Timer status:"))
		runSystemctl("--no-pager", "status", fmt.Sprintf("%s-backup.timer", backupName))
	},
}

func runSystemctl(args ...string) {
	allArgs := append([]string{"--user"}, args...)
	c := exec.Command("systemctl", allArgs...)
	output, err := c.CombinedOutput()
	if err != nil {
		fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error running systemctl: %v\n%s", err, string(output))))
		return
	}
	fmt.Print(string(output))
}
