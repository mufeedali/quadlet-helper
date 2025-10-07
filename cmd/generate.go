package cmd

import (
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate example files",
	Long:  `The generate command helps create example files from your existing configurations.`,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
