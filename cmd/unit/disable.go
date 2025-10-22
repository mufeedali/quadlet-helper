package unit

import (
	"fmt"
	"os"
	"regexp"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var disableCmd = &cobra.Command{
	Use:               "disable <unit-name>...",
	Short:             "Disable one or more quadlet units from starting on boot",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		containersPath := viper.GetString("containers-path")
		realContainersPath := shared.ResolveContainersDir(containersPath)

		var anyChanged bool
		var failures int

		for _, unitName := range args {
			quadletFile, err := findQuadletFile(realContainersPath, unitName)
			if err != nil {
				fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error: %v", err)))
				failures++
				continue
			}

			fmt.Println("  Found: " + shared.FilePathStyle.Render(quadletFile))

			content, err := os.ReadFile(quadletFile)
			if err != nil {
				fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error reading file: %v", err)))
				failures++
				continue
			}

			newContent, changed := removeInstallSection(string(content))
			if !changed {
				fmt.Println(shared.WarningStyle.Render(fmt.Sprintf("Unit %s is not enabled, nothing to do.", unitName)))
				continue
			}

			err = os.WriteFile(quadletFile, []byte(newContent), 0644)
			if err != nil {
				fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error writing file: %v", err)))
				failures++
				continue
			}

			fmt.Println(shared.CheckMark + " Removed [Install] section from " + shared.FilePathStyle.Render(quadletFile))
			anyChanged = true
			fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully disabled %s", unitName)))
		}

		if anyChanged {
			fmt.Println("  Running systemctl --user daemon-reload...")
			if _, err := systemd.DaemonReload(); err != nil {
				fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error running daemon-reload: %v", err)))
				os.Exit(1)
			}
			fmt.Println(shared.CheckMark + " Daemon reloaded.")
		}

		if failures > 0 {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("✗ %d unit(s) failed to disable.", failures)))
			os.Exit(1)
		}

		// If we got here and no failures, print an overall success line
		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully processed %d unit(s)", len(args))))
	},
}

func removeInstallSection(content string) (string, bool) {
	re := regexp.MustCompile(`(?m)^\s*\[Install\][\s\S]*?(\n\s*\[|$)`)
	if !re.MatchString(content) {
		return content, false
	}
	newContent := re.ReplaceAllString(content, "$1")
	return newContent, true
}
