package backup

import (
	"fmt"

	internalbackup "github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:               "cleanup [backup-name]",
	Short:             "Run retention cleanup (used by systemd)",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getBackupNameCompletions(),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupName := args[0]

		config, err := loadBackupConfig(backupName)
		if err != nil {
			return err
		}

		if err := internalbackup.Cleanup(config); err != nil {
			return cmdutil.Wrap(err, "cleanup failed")
		}

		fmt.Println(shared.SuccessStyle.Render("✓ Cleanup completed"))
		return nil
	},
}
