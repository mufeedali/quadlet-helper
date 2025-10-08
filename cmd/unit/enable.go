package unit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var enableCmd = &cobra.Command{
	Use:               "enable <unit-name>",
	Short:             "Enable a quadlet unit to start on boot",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		unitName := args[0]
		containersPath := viper.GetString("containers-path")
		realContainersPath := shared.ResolveContainersDir(containersPath)

		quadletFile, err := findQuadletFile(realContainersPath, unitName)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error: %v", err)))
			os.Exit(1)
		}

		fmt.Println("  Found: " + shared.FilePathStyle.Render(quadletFile))

		content, err := os.ReadFile(quadletFile)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error reading file: %v", err)))
			os.Exit(1)
		}

		newContent, changed := addInstallSection(string(content))
		if !changed {
			fmt.Println(shared.WarningStyle.Render("Unit is already enabled."))
			return
		}

		err = os.WriteFile(quadletFile, []byte(newContent), 0644)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error writing file: %v", err)))
			os.Exit(1)
		}

		fmt.Println(shared.CheckMark + " Added [Install] section to " + shared.FilePathStyle.Render(quadletFile))

		runDaemonReload()

		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully enabled %s", unitName)))
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
WantedBy=default.target
`
	return content + installSection, true
}

func runDaemonReload() {
	fmt.Println("  Running systemctl --user daemon-reload...")
	c := exec.Command("systemctl", "--user", "daemon-reload")
	output, err := c.CombinedOutput()
	if err != nil {
		fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error running daemon-reload: %v\n%s", err, string(output))))
		os.Exit(1)
	}
	fmt.Println(shared.CheckMark + " Daemon reloaded.")
}
