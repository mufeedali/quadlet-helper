package backup

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:               "edit [backup-name]",
	Short:             "Edit backup configuration",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	Run: func(cmd *cobra.Command, args []string) {
		backupName := args[0]

		// Get config path
		configPath, err := backup.GetConfigPath(backupName)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error: %v", err)))
			os.Exit(1)
		}

		// Check if config exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Backup '%s' does not exist", backupName)))
			os.Exit(1)
		}

		// Get editor from environment or use default
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Editing backup: %s", backupName)))
		fmt.Println(shared.FilePathStyle.Render(configPath))
		fmt.Println()

		// Open editor
		editorCmd := exec.Command(editor, configPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		if err := editorCmd.Run(); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error opening editor: %v", err)))
			os.Exit(1)
		}

		// Validate the edited config
		config, err := backup.LoadConfig(backupName)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error loading edited config: %v", err)))
			fmt.Println("Please fix the configuration file and try again.")
			os.Exit(1)
		}

		if err := config.Validate(); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Configuration validation failed: %v", err)))
			fmt.Println("Please fix the configuration file and try again.")
			os.Exit(1)
		}

		fmt.Println(shared.SuccessStyle.Render("✓ Configuration updated successfully"))

		// Check if backup is installed
		timerFilePath, _ := backup.GetTimerFilePath(backupName)
		if _, err := os.Stat(timerFilePath); err == nil {
			fmt.Println("\nNote: The backup is installed. You may need to reinstall it for changes to take effect:")
			fmt.Printf("  qh backup uninstall %s\n", backupName)
			fmt.Printf("  qh backup install %s\n", backupName)
		}
	},
}
