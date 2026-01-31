package generate

import (
	"github.com/spf13/cobra"
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate example files",
	Long:  `The generate command helps create example files from your existing configurations.`,
}

func init() {
	GenerateCmd.PersistentFlags().Bool("hook", false, "Exit 1 if any files were changed (for git hooks)")
	GenerateCmd.AddCommand(envCmd)
	GenerateCmd.AddCommand(traefikCmd)
}
