package unit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var enableCmd = &cobra.Command{
	Use:               "enable <unit-name>...",
	Short:             "Enable one or more quadlet units to start on boot",
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

			newContent, changed := addInstallSection(string(content))
			if !changed {
				fmt.Println(shared.WarningStyle.Render(fmt.Sprintf("Unit %s is already enabled.", unitName)))
				continue
			}

			err = os.WriteFile(quadletFile, []byte(newContent), 0644)
			if err != nil {
				fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error writing file: %v", err)))
				failures++
				continue
			}

			fmt.Println(shared.CheckMark + " Added [Install] section to " + shared.FilePathStyle.Render(quadletFile))
			anyChanged = true
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
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("✗ %d unit(s) failed to enable.", failures)))
			os.Exit(1)
		}

		// Report overall success (if nothing changed, individual warnings were already printed)
		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully processed %d unit(s)", len(args))))
	},
}

func findQuadletFile(dir, unitName string) (string, error) {
	var foundPath string
	err := shared.WalkWithSymlinks(dir, func(path string, info os.FileInfo) error {
		if !info.IsDir() {
			baseName := filepath.Base(path)
			ext := filepath.Ext(baseName)
			if baseName[:len(baseName)-len(ext)] == unitName {
				foundPath = path
				return filepath.SkipDir // Stop searching
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if foundPath == "" {
		return "", fmt.Errorf("quadlet file for unit '%s' not found", unitName)
	}
	return foundPath, nil
}

func addInstallSection(content string) (string, bool) {
	if strings.Contains(content, "[Install]") {
		return content, false
	}
	installSection := `
[Install]
WantedBy=multi-user.target default.target
`
	return content + installSection, true
}
