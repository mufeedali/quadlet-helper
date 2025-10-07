package cloudflare

import (
	"github.com/spf13/cobra"
)

var CloudflareCmd = &cobra.Command{
	Use:   "cloudflare",
	Short: "Manage the Cloudflare IP updater service",
	Long:  `This command helps manage the Cloudflare IP updater service, including installation, uninstallation, and running the update process.`,
}

func init() {
	CloudflareCmd.AddCommand(installCmd)
	CloudflareCmd.AddCommand(runCmd)
	CloudflareCmd.AddCommand(uninstallCmd)
}
