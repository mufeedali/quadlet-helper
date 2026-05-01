package unit

import (
	"bytes"
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/mufeedali/quadlet-helper/internal/quadlet"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func printTable(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				if n := utf8.RuneCountInString(cell); n > widths[i] {
					widths[i] = n
				}
			}
		}
	}

	// border wraps each box-drawing character in grey; cell text is left unstyled.
	border := func(s string) string {
		if os.Getenv("NO_COLOR") != "" {
			return s
		}
		return "\x1b[38;5;238m" + s + "\x1b[0m"
	}

	hline := func(left, mid, right string) {
		segs := make([]string, len(widths))
		for i, w := range widths {
			segs[i] = strings.Repeat("─", w+2)
		}
		fmt.Println(border(left + strings.Join(segs, mid) + right))
	}

	printRow := func(cells []string, isHeader bool) {
		var b strings.Builder
		for i, cell := range cells {
			pad := strings.Repeat(" ", widths[i]-utf8.RuneCountInString(cell))
			if isHeader {
				cell = shared.TitleStyle.Render(cell)
			}
			b.WriteString(border("│") + " " + cell + pad + " ")
		}
		fmt.Println(b.String() + border("│"))
	}

	hline("╭", "┬", "╮")
	printRow(headers, true)
	hline("├", "┼", "┤")
	for _, row := range rows {
		printRow(row, false)
	}
	hline("╰", "┴", "╯")
}

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

		printTable([]string{"Unit", "Type", "Boot", "Running"}, tableRows)
		return nil
	},
}

