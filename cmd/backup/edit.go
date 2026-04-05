package backup

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:               "edit [backup-name]",
	Short:             "Edit backup configuration",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupName := args[0]

		configPath, err := backup.GetConfigPath(backupName)
		if err != nil {
			return cmdutil.Wrap(err, "resolving config path")
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return cmdutil.Errorf("backup %q does not exist", backupName)
		}

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
			return cmdutil.Wrap(err, "opening editor")
		}

		config, err := loadBackupConfig(backupName)
		if err != nil {
			return cmdutil.Errorf("%v\nPlease fix the configuration file and try again.", err)
		}

		if err := config.Validate(); err != nil {
			return cmdutil.Errorf("configuration validation failed: %v\nPlease fix the configuration file and try again.", err)
		}

		fmt.Println(shared.SuccessStyle.Render("✓ Configuration updated successfully"))

		if isInstalledBackup(backupName) {
			fmt.Println("\nNote: The backup is installed. You may need to reinstall it for changes to take effect:")
			fmt.Printf("  qh backup uninstall %s\n", backupName)
			fmt.Printf("  qh backup install %s\n", backupName)
		}

		return nil
	},
}
