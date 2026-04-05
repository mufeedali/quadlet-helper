package backup

import (
	"fmt"

	internalbackup "github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all backup configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(shared.TitleStyle.Render("Backup Configurations"))
		fmt.Println()

		configs, err := internalbackup.ListConfigs()
		if err != nil {
			return cmdutil.Wrap(err, "listing configs")
		}

		if len(configs) == 0 {
			fmt.Println("No backup configurations found.")
			fmt.Println("\nCreate a new backup with:")
			fmt.Println("  qh backup create")
			return nil
		}

		for _, name := range configs {
			config, err := loadBackupConfig(name)
			if err != nil {
				fmt.Printf("%s %s (error loading: %v)\n", shared.ErrorStyle.Render("✗"), name, err)
				continue
			}

			installed := isInstalledBackup(name)

			active := false
			timerName := internalbackup.BackupTimerName(name)
			if installed {
				active, err = systemd.IsActive(timerName)
				if err != nil {
					fmt.Printf("%s %s (error checking status: %v)\n", shared.WarningStyle.Render("!"), name, err)
					continue
				}
			}

			status := "not installed"
			statusStyle := shared.WarningStyle
			if installed {
				if active {
					status = "active"
					statusStyle = shared.SuccessStyle
				} else {
					status = "inactive"
					statusStyle = shared.ErrorStyle
				}
			}

			fmt.Printf("%s %s [%s] - %s → %s\n",
				statusStyle.Render("●"),
				shared.TitleStyle.Render(name),
				statusStyle.Render(status),
				config.Type,
				config.GetDestination())
			fmt.Printf("  Schedule: %s\n", config.Schedule)

			if active {
				// Show next run time
				if output, err := systemd.Show(timerName, "NextElapseUSecRealtime"); err == nil && len(output) > 0 {
					fmt.Printf("  Next run: %s", output)
				}
			}

			fmt.Println()
		}

		return nil
	},
}
