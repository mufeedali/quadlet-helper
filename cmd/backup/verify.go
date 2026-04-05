package backup

import (
	"fmt"

	internalbackup "github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:               "verify [backup-name]",
	Short:             "Verify backup integrity",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupName := args[0]

		fmt.Println("Verifying backup: " + shared.TitleStyle.Render(backupName))

		config, err := loadBackupConfig(backupName)
		if err != nil {
			return err
		}

		result, err := internalbackup.Verify(config)
		if err != nil {
			return cmdutil.Wrap(err, "verification error")
		}

		if !result.Success {
			if result.Details != "" {
				return cmdutil.Errorf("verification failed: %s\n\nDetails:\n%s", result.Message, result.Details)
			}
			return cmdutil.Errorf("verification failed: %s", result.Message)
		}

		fmt.Println(shared.SuccessStyle.Render("✓ Verification successful"))
		fmt.Println(result.Message)
		if result.Details != "" {
			fmt.Println("\nDetails:")
			fmt.Println(result.Details)
		}

		return nil
	},
}
