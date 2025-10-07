package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var cloudflareUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Cloudflare IP updater service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(titleStyle.Render("Uninstalling Cloudflare IP Updater..."))

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error finding home directory: %v", err)))
			os.Exit(1)
		}
		systemdUserDir := filepath.Join(home, ".config", "systemd", "user")

		runSystemctl("stop", "cloudflare-ip-updater.timer")
		runSystemctl("disable", "cloudflare-ip-updater.timer")
		runSystemctl("stop", "cloudflare-ip-updater.service")

		serviceFile := filepath.Join(systemdUserDir, "cloudflare-ip-updater.service")
		if err := os.Remove(serviceFile); err == nil {
			fmt.Println(checkMark + " Removed " + filePathStyle.Render(serviceFile))
		} else if !os.IsNotExist(err) {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error removing service file: %v", err)))
		}

		timerFile := filepath.Join(systemdUserDir, "cloudflare-ip-updater.timer")
		if err := os.Remove(timerFile); err == nil {
			fmt.Println(checkMark + " Removed " + filePathStyle.Render(timerFile))
		} else if !os.IsNotExist(err) {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error removing timer file: %v", err)))
		}

		runSystemctl("daemon-reload")

		fmt.Println(successStyle.Render("\n✓ Uninstallation complete!"))
	},
}

func init() {
	cloudflareCmd.AddCommand(cloudflareUninstallCmd)
}
