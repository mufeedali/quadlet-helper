package unit

import (
	"github.com/spf13/cobra"
)

var UnitCmd = &cobra.Command{
	Use:   "unit",
	Short: "Manage quadlet unit files",
	Long:  `The unit command provides tools for creating, managing, and interacting with quadlet unit files.`,
}

func init() {
	UnitCmd.AddCommand(disableCmd)
	UnitCmd.AddCommand(enableCmd)
	UnitCmd.AddCommand(listCmd)
	UnitCmd.AddCommand(logsCmd)
	UnitCmd.AddCommand(restartCmd)
	UnitCmd.AddCommand(startCmd)
	UnitCmd.AddCommand(statusCmd)
	UnitCmd.AddCommand(stopCmd)
	UnitCmd.AddCommand(validateCmd)
}
