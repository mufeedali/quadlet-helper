package cloudflare

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

const serviceTemplate = `[Unit]
Description=Update Cloudflare IP ranges in Traefik config
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
ExecStart=%s cloudflare run

Restart=no

StandardOutput=journal
StandardError=journal
`

const timerTemplate = `[Unit]
Description=Update Cloudflare IPs weekly
Requires=cloudflare-ip-updater.service

[Timer]
OnCalendar=Sun *-*-* 03:00:00
OnBootSec=5min
Persistent=true

[Install]
WantedBy=timers.target
`

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Cloudflare IP updater service",
	Run: func(c *cobra.Command, args []string) {
		fmt.Println(shared.TitleStyle.Render("Installing Cloudflare IP Updater..."))

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error finding home directory: %v", err)))
			os.Exit(1)
		}
		systemdUserDir := filepath.Join(home, ".config", "systemd", "user")
		if err := os.MkdirAll(systemdUserDir, 0755); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error creating systemd directory: %v", err)))
			os.Exit(1)
		}

		executablePath, err := os.Executable()
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error finding executable path: %v", err)))
			os.Exit(1)
		}

		serviceContent := fmt.Sprintf(serviceTemplate, executablePath)
		serviceFilePath := filepath.Join(systemdUserDir, "cloudflare-ip-updater.service")
		err = os.WriteFile(serviceFilePath, []byte(serviceContent), 0644)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error writing service file: %v", err)))
			os.Exit(1)
		}
		fmt.Println(shared.CheckMark + " Created " + shared.FilePathStyle.Render(serviceFilePath))

		timerFilePath := filepath.Join(systemdUserDir, "cloudflare-ip-updater.timer")
		err = os.WriteFile(timerFilePath, []byte(timerTemplate), 0644)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error writing timer file: %v", err)))
			os.Exit(1)
		}
		fmt.Println(shared.CheckMark + " Created " + shared.FilePathStyle.Render(timerFilePath))

		runSystemctl("daemon-reload")
		runSystemctl("enable", "cloudflare-ip-updater.timer")
		runSystemctl("start", "cloudflare-ip-updater.timer")

		fmt.Println(shared.SuccessStyle.Render("\n✓ Installation complete!"))
		fmt.Println(shared.TitleStyle.Render("Timer status:"))
		runSystemctl("--no-pager", "status", "cloudflare-ip-updater.timer")
	},
}

func runSystemctl(args ...string) {
	allArgs := append([]string{"--user"}, args...)
	c := exec.Command("systemctl", allArgs...)
	output, err := c.CombinedOutput()
	if err != nil {
		fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error running systemctl %s: %v\n%s", strings.Join(args, " "), err, string(output))))
		return
	}
	fmt.Print(string(output))
}
