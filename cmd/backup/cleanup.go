package backup

import (
	"fmt"
	"os"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:               "cleanup [backup-name]",
	Short:             "Run retention cleanup (used by systemd)",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	Run: func(cmd *cobra.Command, args []string) {
		backupName := args[0]

		// Load config
		config, err := backup.LoadConfig(backupName)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error loading config: %v", err)))
			os.Exit(1)
		}

		// Run cleanup
		if err := backup.Cleanup(config); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Cleanup failed: %v", err)))
			os.Exit(1)
		}

		fmt.Println(shared.SuccessStyle.Render("✓ Cleanup completed"))
	},
}
