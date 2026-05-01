package unit

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/quadlet"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var errValidationFailed = errors.New("validation failed")

var validateCmd = &cobra.Command{
	Use:   "validate [unit-name...]",
	Short: "Validate quadlet unit file(s)",
	Long: `This command runs systemd's own generator and verification tools
to check for errors in quadlet files before they are installed.`,
	ValidArgsFunction: unitCompletionFunc,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			containersPath := viper.GetString("containers-path")
			realContainersPath := shared.ResolveContainersDir(containersPath)

			units, err := quadlet.List(realContainersPath)
			if err != nil {
				return err
			}
			index := make(map[string]*quadlet.Unit, len(units))
			for i, u := range units {
				index[u.BaseName()] = &units[i]
			}

			var failures int
			for _, unitName := range args {
				fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Validating %s...", unitName)))
				u, ok := index[unitName]
				if !ok {
					fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("✗ Unit %q not found", unitName)))
					failures++
					continue
				}
				ok2, output := runValidate(u.UnitName)
				if !ok2 {
					fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("✗ Validation failed for %s:", unitName)))
					fmt.Println(output)
					failures++
				} else {
					fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ %s is valid.", unitName)))
				}
			}
			if failures > 0 {
				return cmdutil.Errorf("validation failed for %d unit(s)", failures)
			}
			return nil
		}
		return validateAllUnits()
	},
}

// runValidate runs systemd-analyze verify on a service unit name and returns
// whether it passed and any output.
func runValidate(serviceName string) (bool, string) {
	cmd := exec.Command("systemd-analyze", "--user", "--generators=true", "verify", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, string(output)
	}
	return true, ""
}

func validateAllUnits() error {
	containersPath := viper.GetString("containers-path")
	realContainersPath := shared.ResolveContainersDir(containersPath)

	fmt.Println(shared.TitleStyle.Render("Validating all Quadlet Units in " + shared.FilePathStyle.Render(realContainersPath) + "\n"))

	units, err := quadlet.List(realContainersPath)
	if err != nil {
		return err
	}

	if len(units) == 0 {
		fmt.Println(shared.WarningStyle.Render("No quadlet files found."))
		return nil
	}

	var validCount, failedCount int
	for _, u := range units {
		ok, output := runValidate(u.UnitName)
		if ok {
			fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("  ✓ %s", u.BaseName())))
			validCount++
		} else {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("  ✗ %s", u.BaseName())))
			for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
				fmt.Printf("    %s\n", line)
			}
			failedCount++
		}
	}

	summary := fmt.Sprintf("\nValidation complete. (%d valid, %d failed)", validCount, failedCount)
	if failedCount > 0 {
		fmt.Println(shared.ErrorStyle.Render(summary))
		return errValidationFailed
	}
	fmt.Println(shared.SuccessStyle.Render(summary))
	return nil
}
