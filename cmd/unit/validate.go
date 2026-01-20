package unit

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var validateCmd = &cobra.Command{
	Use:   "validate [unit-name...]",
	Short: "Validate quadlet unit file(s)",
	Long: `This command runs systemd's own generator and verification tools
to check for errors in quadlet files before they are installed.`,
	ValidArgsFunction: unitCompletionFunc,
	Run: func(cmd *cobra.Command, args []string) {
		// If one or more unit names are provided, validate each individually.
		// If no args are provided, validate all units (existing behavior).
		if len(args) > 0 {
			var failures int
			for _, unitName := range args {
				fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Validating %s...", unitName)))
				ok, output := validateUnit(unitName)
				if !ok {
					fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("✗ Validation failed for %s:", unitName)))
					fmt.Println(output)
					failures++
				} else {
					fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ %s is valid.", unitName)))
				}
			}
			if failures > 0 {
				os.Exit(1)
			}
		} else {
			if !validateAllUnits() {
				os.Exit(1)
			}
		}
	},
}

// validateUnit checks a single unit and returns its validity and any output.
// It does not print anything itself.
func validateUnit(unitName string) (bool, string) {
	containersPath := viper.GetString("containers-path")
	realContainersPath := shared.ResolveContainersDir(containersPath)
	quadletFile, err := findQuadletFile(realContainersPath, unitName)
	if err != nil {
		return false, fmt.Sprintf("Error finding unit: %v", err)
	}

	serviceName := getServiceNameFromExtension(unitName, filepath.Ext(quadletFile))

	cmd := exec.Command("systemd-analyze", "--user", "--generators=true", "verify", serviceName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return false, string(output)
	}

	return true, ""
}

func validateAllUnits() bool {
	containersPath := viper.GetString("containers-path")
	realContainersPath := shared.ResolveContainersDir(containersPath)

	fmt.Println(shared.TitleStyle.Render("Validating all Quadlet Units in " + shared.FilePathStyle.Render(realContainersPath) + "\n"))

	var validCount, failedCount int
	foundAny := false

	err := shared.WalkWithSymlinks(realContainersPath, func(path string, d fs.DirEntry) error {
		if !d.IsDir() {
			ext := filepath.Ext(d.Name())
			if isQuadletUnit(ext) {
				foundAny = true
				unitName := strings.TrimSuffix(d.Name(), ext)
				ok, output := validateUnit(unitName)
				if ok {
					fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("  ✓ %s", unitName)))
					validCount++
				} else {
					fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("  ✗ %s", unitName)))
					// Indent the error output for clarity
					for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
						fmt.Printf("    %s\n", line)
					}
					failedCount++
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("\nError walking directory: %v", err)))
		return false
	}

	if !foundAny {
		fmt.Println(shared.WarningStyle.Render("No quadlet files found."))
	}

	summary := fmt.Sprintf("\nValidation complete. (%d valid, %d failed)", validCount, failedCount)
	if failedCount > 0 {
		fmt.Println(shared.ErrorStyle.Render(summary))
	} else {
		fmt.Println(shared.SuccessStyle.Render(summary))
	}

	return failedCount == 0
}
