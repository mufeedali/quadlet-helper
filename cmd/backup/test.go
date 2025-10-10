package backup

import (
	"fmt"
	"os"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:               "test [backup-name]",
	Short:             "Test backup configuration (dry-run)",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	Run: func(cmd *cobra.Command, args []string) {
		backupName := args[0]

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Testing backup: %s (dry-run)", backupName)))
		fmt.Println()

		// Load config
		config, err := backup.LoadConfig(backupName)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error loading config: %v", err)))
			os.Exit(1)
		}

		// Validate config
		if err := config.Validate(); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Configuration validation failed: %v", err)))
			os.Exit(1)
		}

		fmt.Println(shared.SuccessStyle.Render("✓ Configuration is valid"))
		fmt.Println()

		// Run dry-run backup
		result, err := backup.Run(config, true)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Dry-run failed: %v", err)))
			fmt.Println("\nOutput:")
			fmt.Println(result.Output)
			os.Exit(1)
		}

		fmt.Println(result.Output)
		fmt.Println()
		fmt.Println(shared.SuccessStyle.Render("✓ Dry-run completed successfully"))
		fmt.Println("\nThe backup configuration is working correctly.")
		fmt.Println("To run the actual backup:")
		fmt.Printf("  qh backup run %s\n", backupName)
	},
}
