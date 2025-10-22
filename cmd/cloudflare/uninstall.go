package cloudflare

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Cloudflare IP updater service",
	Run: func(c *cobra.Command, args []string) {
		fmt.Println(shared.TitleStyle.Render("Uninstalling Cloudflare IP Updater..."))

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error finding home directory: %v", err)))
			os.Exit(1)
		}
		systemdUserDir := filepath.Join(home, ".config", "systemd", "user")

		if _, err := systemd.Stop("cloudflare-ip-updater.timer"); err != nil {
			fmt.Println(shared.WarningStyle.Render(fmt.Sprintf("Could not stop timer: %v", err)))
		}
		if _, err := systemd.Disable("cloudflare-ip-updater.timer"); err != nil {
			fmt.Println(shared.WarningStyle.Render(fmt.Sprintf("Could not disable timer: %v", err)))
		}
		if _, err := systemd.Stop("cloudflare-ip-updater.service"); err != nil {
			fmt.Println(shared.WarningStyle.Render(fmt.Sprintf("Could not stop service: %v", err)))
		}

		serviceFile := filepath.Join(systemdUserDir, "cloudflare-ip-updater.service")
		if err := os.Remove(serviceFile); err == nil {
			fmt.Println(shared.CheckMark + " Removed " + shared.FilePathStyle.Render(serviceFile))
		} else if !os.IsNotExist(err) {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error removing service file: %v", err)))
		}

		timerFile := filepath.Join(systemdUserDir, "cloudflare-ip-updater.timer")
		if err := os.Remove(timerFile); err == nil {
			fmt.Println(shared.CheckMark + " Removed " + shared.FilePathStyle.Render(timerFile))
		} else if !os.IsNotExist(err) {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error removing timer file: %v", err)))
		}

		if _, err := systemd.DaemonReload(); err != nil {
			fmt.Println(shared.WarningStyle.Render(fmt.Sprintf("Could not reload systemd: %v", err)))
		}

		fmt.Println(shared.SuccessStyle.Render("\n✓ Uninstallation complete!"))
	},
}
