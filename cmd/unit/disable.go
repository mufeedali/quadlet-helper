package unit

import (
	"fmt"
	"os"
	"regexp"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var disableCmd = &cobra.Command{
	Use:               "disable <unit-name>",
	Short:             "Disable a quadlet unit from starting on boot",
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

		newContent, changed := removeInstallSection(string(content))
		if !changed {
			fmt.Println(shared.WarningStyle.Render("Unit is not enabled, nothing to do."))
			return
		}

		err = os.WriteFile(quadletFile, []byte(newContent), 0644)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error writing file: %v", err)))
			os.Exit(1)
		}

		fmt.Println(shared.CheckMark + " Removed [Install] section from " + shared.FilePathStyle.Render(quadletFile))

		runDaemonReload()

		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully disabled %s", unitName)))
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
