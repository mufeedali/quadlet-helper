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
	GenerateCmd.AddCommand(envCmd)
	GenerateCmd.AddCommand(traefikCmd)
}
