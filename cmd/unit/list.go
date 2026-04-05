package unit

import (
	"bytes"
	"cmp"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		containersPath := viper.GetString("containers-path")
		realContainersPath := shared.ResolveContainersDir(containersPath)

		type unitInfo struct {
			name        string
			serviceName string
			unitType    string
			enabled     string
			status      string
		}

		var units []*unitInfo
		var unitsToCheck []*unitInfo
		foundAny := false

		err := shared.WalkWithSymlinks(realContainersPath, func(path string, d fs.DirEntry) error {
			if d.IsDir() {
				// Skip dir
				return nil
			}

			ext := filepath.Ext(d.Name())
			if !isQuadletUnit(ext) {
				// Skip non-quadlet files (like maybe READMEs)
				return nil
			}

			foundAny = true
			unitName := strings.TrimSuffix(d.Name(), ext)
			unitType := strings.TrimPrefix(ext, ".")
			serviceName := getServiceNameFromExtension(unitName, ext)

			// Check if unit is enabled by reading the file
			enabledStatus := "-"
			if unitType != "network" {
				enabledStatus = "✗"
				content, err := os.ReadFile(path)
				if err == nil && bytes.Contains(content, []byte("[Install]")) {
					enabledStatus = "✓"
				}
			}

			// Status will be set later in a batch call
			activeStatus := ""

			// Don't show status for units that are typically 'oneshot'
			if unitType == "volume" || unitType == "image" || unitType == "build" {
				enabledStatus = "-"
				activeStatus = "-"
			}

			u := &unitInfo{
				name:        unitName,
				serviceName: serviceName,
				unitType:    unitType,
				enabled:     enabledStatus,
				status:      activeStatus,
			}
			units = append(units, u)

			if activeStatus != "-" {
				unitsToCheck = append(unitsToCheck, u)
			}
			return nil
		})

		if err != nil {
			return err
		}

		if !foundAny {
			fmt.Println(shared.WarningStyle.Render("No quadlet files found."))
			return nil
		}

		// Batch check active status for all units in a single systemctl call
		if len(unitsToCheck) > 0 {
			serviceNames := make([]string, len(unitsToCheck))
			for i, u := range unitsToCheck {
				serviceNames[i] = u.serviceName
			}

			activeStatuses, err := systemd.IsActiveMultiple(serviceNames)
			if err != nil {
				return err
			}

			for i, active := range activeStatuses {
				if active {
					unitsToCheck[i].status = "✓"
				} else {
					unitsToCheck[i].status = "✗"
				}
			}
		}

		// Sort units before processing
		slices.SortFunc(units, func(a, b *unitInfo) int {
			if a.unitType != b.unitType {
				return cmp.Compare(a.unitType, b.unitType) // Group order
			}
			return cmp.Compare(a.name, b.name) // Sort within group
		})

		// Now build rows in correct order
		rows := [][]string{}
		for _, u := range units {
			rows = append(rows, []string{u.name, u.unitType, u.enabled, u.status})
		}

		// Build the table with lipgloss
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
		return nil
	},
}
