package backup

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all backup configurations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(shared.TitleStyle.Render("Backup Configurations"))
		fmt.Println()

		configs, err := backup.ListConfigs()
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error listing configs: %v", err)))
			os.Exit(1)
		}

		if len(configs) == 0 {
			fmt.Println("No backup configurations found.")
			fmt.Println("\nCreate a new backup with:")
			fmt.Println("  qh backup create")
			return
		}

		for _, name := range configs {
			config, err := backup.LoadConfig(name)
			if err != nil {
				fmt.Printf("%s %s (error loading: %v)\n", shared.ErrorStyle.Render("✗"), name, err)
				continue
			}

			// Check if service is installed
			timerFilePath, _ := backup.GetTimerFilePath(name)
			installed := fileExists(timerFilePath)

			// Check if timer is active
			active := false
			if installed {
				cmd := exec.Command("systemctl", "--user", "is-active", fmt.Sprintf("%s-backup.timer", name))
				if err := cmd.Run(); err == nil {
					active = true
				}
			}

			// Display status
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
				cmd := exec.Command("systemctl", "--user", "show", fmt.Sprintf("%s-backup.timer", name), "--property=NextElapseUSecRealtime", "--value")
				if output, err := cmd.Output(); err == nil && len(output) > 0 {
					fmt.Printf("  Next run: %s", string(output))
				}
			}

			fmt.Println()
		}
	},
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
