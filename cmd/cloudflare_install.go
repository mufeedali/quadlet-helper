package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

var cloudflareInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Cloudflare IP updater service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(titleStyle.Render("Installing Cloudflare IP Updater..."))

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error finding home directory: %v", err)))
			os.Exit(1)
		}
		systemdUserDir := filepath.Join(home, ".config", "systemd", "user")
		os.MkdirAll(systemdUserDir, 0755)

		executablePath, err := os.Executable()
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error finding executable path: %v", err)))
			os.Exit(1)
		}

		serviceContent := fmt.Sprintf(serviceTemplate, executablePath)
		serviceFilePath := filepath.Join(systemdUserDir, "cloudflare-ip-updater.service")
		err = os.WriteFile(serviceFilePath, []byte(serviceContent), 0644)
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error writing service file: %v", err)))
			os.Exit(1)
		}
		fmt.Println(checkMark + " Created " + filePathStyle.Render(serviceFilePath))

		timerFilePath := filepath.Join(systemdUserDir, "cloudflare-ip-updater.timer")
		err = os.WriteFile(timerFilePath, []byte(timerTemplate), 0644)
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error writing timer file: %v", err)))
			os.Exit(1)
		}
		fmt.Println(checkMark + " Created " + filePathStyle.Render(timerFilePath))

		runSystemctl("daemon-reload")
		runSystemctl("enable", "cloudflare-ip-updater.timer")
		runSystemctl("start", "cloudflare-ip-updater.timer")

		fmt.Println(successStyle.Render("\n✓ Installation complete!"))
		fmt.Println(titleStyle.Render("Timer status:"))
		runSystemctl("--no-pager", "status", "cloudflare-ip-updater.timer")
	},
}

func runSystemctl(args ...string) {
	allArgs := append([]string{"--user"}, args...)
	cmd := exec.Command("systemctl", allArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error running systemctl %s: %v\n%s", strings.Join(args, " "), err, string(output))))
		return
	}
	fmt.Print(string(output))
}

func init() {
	cloudflareCmd.AddCommand(cloudflareInstallCmd)
}
