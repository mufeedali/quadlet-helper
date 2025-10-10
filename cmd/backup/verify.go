package backup

import (
	"fmt"
	"os"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:               "verify [backup-name]",
	Short:             "Verify backup integrity",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	Run: func(cmd *cobra.Command, args []string) {
		backupName := args[0]

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Verifying backup: %s", backupName)))
		fmt.Println()

		// Load config
		config, err := backup.LoadConfig(backupName)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error loading config: %v", err)))
			os.Exit(1)
		}

		// Verify backup
		result, err := backup.Verify(config)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Verification error: %v", err)))
			os.Exit(1)
		}

		if !result.Success {
			fmt.Println(shared.ErrorStyle.Render("✗ Verification failed"))
			fmt.Println(shared.ErrorStyle.Render(result.Message))
			if result.Details != "" {
				fmt.Println("\nDetails:")
				fmt.Println(result.Details)
			}
			os.Exit(1)
		}

		fmt.Println(shared.SuccessStyle.Render("✓ Verification successful"))
		fmt.Println(result.Message)
		if result.Details != "" {
			fmt.Println("\nDetails:")
			fmt.Println(result.Details)
		}
	},
}
