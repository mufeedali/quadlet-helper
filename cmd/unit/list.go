package unit

import (
	"bytes"
	"cmp"
	"fmt"
	"os"
	"slices"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mufeedali/quadlet-helper/internal/quadlet"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all quadlet units and their status",
	RunE: func(cmd *cobra.Command, args []string) error {
		containersPath := viper.GetString("containers-path")
		realContainersPath := shared.ResolveContainersDir(containersPath)

		units, err := quadlet.List(realContainersPath)
		if err != nil {
			return err
		}

		if len(units) == 0 {
			fmt.Println(shared.WarningStyle.Render("No quadlet files found."))
			return nil
		}

		type row struct {
			name     string
			unitType string
			enabled  string
			status   string
		}

		rows := make([]row, 0, len(units))
		for _, u := range units {
			unitType := u.UnitType()

			enabledStatus := "-"
			if unitType != "network" && unitType != "volume" && unitType != "image" && unitType != "build" {
				enabledStatus = "✗"
				content, err := os.ReadFile(u.Path)
				if err == nil && bytes.Contains(content, []byte("[Install]")) {
					enabledStatus = "✓"
				}
			}

			activeStatus := "-"
			if unitType != "volume" && unitType != "image" && unitType != "build" {
				if u.IsActive() {
					activeStatus = "✓"
				} else {
					activeStatus = "✗"
				}
			}

			rows = append(rows, row{
				name:     u.BaseName(),
				unitType: unitType,
				enabled:  enabledStatus,
				status:   activeStatus,
			})
		}

		slices.SortFunc(rows, func(a, b row) int {
			if a.unitType != b.unitType {
				return cmp.Compare(a.unitType, b.unitType)
			}
			return cmp.Compare(a.name, b.name)
		})

		tableRows := make([][]string, len(rows))
		for i, r := range rows {
			tableRows[i] = []string{r.name, r.unitType, r.enabled, r.status}
		}

		re := lipgloss.NewRenderer(os.Stdout)

		var (
			CellStyle   = re.NewStyle().Padding(0, 1)
			BorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
		)

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
			Rows(tableRows...)

		fmt.Println(t)
		return nil
	},
}
