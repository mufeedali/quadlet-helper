package cloudflare

import (
	"fmt"
	"os"

	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
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
	RunE: func(c *cobra.Command, args []string) error {
		fmt.Println(shared.TitleStyle.Render("Installing Cloudflare IP Updater..."))

		executablePath, err := os.Executable()
		if err != nil {
			return cmdutil.Wrap(err, "finding executable path")
		}

		paths, err := systemd.InstallUserUnits([]systemd.UserUnitFile{
			{Name: "cloudflare-ip-updater.service", Content: fmt.Sprintf(serviceTemplate, executablePath), Mode: 0644},
			{Name: "cloudflare-ip-updater.timer", Content: timerTemplate, Mode: 0644},
		}, []string{"cloudflare-ip-updater.timer"})
		if err != nil {
			return err
		}
		for _, path := range paths {
			fmt.Println(shared.CheckMark + " Created " + shared.FilePathStyle.Render(path))
		}

		fmt.Println(shared.SuccessStyle.Render("\n✓ Installation complete!"))
		fmt.Println(shared.TitleStyle.Render("Timer status:"))
		output, err := systemd.Status("cloudflare-ip-updater.timer")
		fmt.Println(output)
		if err != nil {
			return cmdutil.Wrap(err, "getting timer status")
		}
		active, err := systemd.IsActive("cloudflare-ip-updater.timer")
		if err != nil {
			return cmdutil.Wrap(err, "checking timer active state")
		}
		if !active {
			return cmdutil.Errorf("timer cloudflare-ip-updater.timer did not become active after installation")
		}
		return nil
	},
}
