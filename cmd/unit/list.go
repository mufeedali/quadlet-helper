package unit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all quadlet units and their status",
	Run: func(cmd *cobra.Command, args []string) {
		containersPath := viper.GetString("containers-path")
		realContainersPath := shared.ResolveContainersDir(containersPath)

		fmt.Println(shared.TitleStyle.Render("Available Quadlet Units in " + shared.FilePathStyle.Render(realContainersPath) + "\n"))

		type unitInfo struct {
			name     string
			unitType string
			enabled  string
			status   string
		}

		var units []unitInfo
		foundAny := false

		err := shared.WalkWithSymlinks(realContainersPath, func(path string, info os.FileInfo) error {
			if !info.IsDir() {
				ext := filepath.Ext(info.Name())
				if isQuadletUnit(ext) {
					foundAny = true
					unitName := strings.TrimSuffix(info.Name(), ext)

					serviceName := unitName + ".service"
					// Adjust service name for specific unit types as per quadlet convention
					switch ext {
					case ".network":
						serviceName = unitName + "-network.service"
					case ".volume":
						serviceName = unitName + "-volume.service"
					case ".pod":
						serviceName = unitName + "-pod.service"
					}

					// Check if unit is enabled by reading the file
					enabledStatus := "✗"
					content, err := os.ReadFile(path)
					if err == nil && strings.Contains(string(content), "[Install]") {
						enabledStatus = "✓"
					}

					activeStatus := getServiceStatus(serviceName)

					// Simplify active status to symbols
					switch activeStatus {
					case "active":
						activeStatus = "✓"
					case "inactive", "failed":
						activeStatus = "✗"
					default:
						output, _ := systemd.Status(serviceName)
						activeStatus = strings.TrimSpace(output)
					} // Clean up unit type - remove the dot prefix
					cleanType := strings.TrimPrefix(ext, ".")

					// Don't show status for units that are typically 'oneshot'
					if ext == ".volume" || ext == ".network" || ext == ".image" || ext == ".build" {
						enabledStatus = "-"
						activeStatus = "-"
					}

					units = append(units, unitInfo{
						name:     unitName,
						unitType: cleanType,
						enabled:  enabledStatus,
						status:   activeStatus,
					})
				}
			}
			return nil
		})

		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error walking directory: %v", err)))
			os.Exit(1)
		}

		if !foundAny {
			fmt.Println(shared.WarningStyle.Render("No quadlet files found."))
			return
		}

		// Build the table with lipgloss
		rows := [][]string{}
		for _, u := range units {
			rows = append(rows, []string{u.name, u.unitType, u.enabled, u.status})
		}

		re := lipgloss.NewRenderer(os.Stdout)

		var (
			CellStyle   = re.NewStyle().Padding(0, 1)
			BorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
		)

		// Style headers with bold
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(BorderStyle).
			StyleFunc(func(row, col int) lipgloss.Style {
				return CellStyle
			}).
			Headers(
				headerStyle.Render("Unit"),
				headerStyle.Render("Type"),
				headerStyle.Render("Boot"),
				headerStyle.Render("Running"),
			).
			Rows(rows...)

		fmt.Println(t)
	},
}

func getServiceStatus(serviceName string) string {
	output, err := systemd.Status(serviceName)
	if err != nil {
		// is-enabled returns non-zero for 'disabled', is-active for 'inactive'
		// We just return the output which is the status word.
	}
	return strings.TrimSpace(string(output))
}
