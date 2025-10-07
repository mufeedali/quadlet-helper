package cmd

import (
	"github.com/spf13/cobra"
)

var cloudflareCmd = &cobra.Command{
	Use:   "cloudflare",
	Short: "Manage the Cloudflare IP updater service",
	Long:  `This command helps manage the Cloudflare IP updater service, including installation, uninstallation, and running the update process.`,
}

func init() {
	rootCmd.AddCommand(cloudflareCmd)
}
