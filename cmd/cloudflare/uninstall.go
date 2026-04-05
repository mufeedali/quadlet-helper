package cloudflare

import (
	"fmt"

	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Cloudflare IP updater service",
	RunE: func(c *cobra.Command, args []string) error {
		fmt.Println(shared.TitleStyle.Render("Uninstalling Cloudflare IP Updater..."))

		result, err := systemd.UninstallUserUnits(
			[]string{"cloudflare-ip-updater.timer", "cloudflare-ip-updater.service"},
			[]string{"cloudflare-ip-updater.timer"},
			[]string{"cloudflare-ip-updater.service", "cloudflare-ip-updater.timer"},
		)
		if err != nil {
			return cmdutil.Wrap(err, "uninstalling cloudflare updater")
		}
		for _, warning := range result.Warnings {
			fmt.Println(shared.WarningStyle.Render("Warning: " + warning.Error()))
		}
		for _, path := range result.RemovedPaths {
			fmt.Println(shared.CheckMark + " Removed " + shared.FilePathStyle.Render(path))
		}

		fmt.Println(shared.SuccessStyle.Render("\n✓ Uninstallation complete!"))
		return nil
	},
}
