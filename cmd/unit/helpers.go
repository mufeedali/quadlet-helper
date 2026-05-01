package unit

import (
	"fmt"
	"os"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/quadlet"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/viper"
)

func loadServices(unitNames []string, types []string) ([]string, error) {
	services, err := resolveServiceNames(unitNames, types)
	if err != nil {
		return nil, cmdutil.Wrap(err, "resolving service names")
	}
	return services, nil
}

func runServiceAction(unitNames []string, types []string, title string, noReload bool, reload bool, action func([]string) (string, error), failure string, success string, hint func([]string) string) error {
	services, err := loadServices(unitNames, types)
	if err != nil {
		return err
	}

	fmt.Println(shared.TitleStyle.Render(fmt.Sprintf(title, strings.Join(services, " "))))
	if reload {
		if !noReload {
			if err := reloadDaemon(); err != nil {
				return err
			}
		} else {
			printSkipReload()
		}
	}

	output, err := action(services)
	if err != nil {
		if hint != nil {
			fmt.Println(shared.InfoMark + " " + hint(services))
		}
		if strings.TrimSpace(output) == "" {
			return cmdutil.Wrap(err, failure)
		}
		return fmt.Errorf("%s: %w\n%s", failure, err, output)
	}

	if strings.TrimSpace(output) != "" {
		fmt.Println(output)
	}
	fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf(success, strings.Join(services, ", "))))
	return nil
}

func updateInstallSections(unitNames []string, actionName string, unchangedMessage func(string) string, apply func(string) (string, bool), changedMessage string) error {
	containersPath := viper.GetString("containers-path")
	realContainersPath := shared.ResolveContainersDir(containersPath)

	allUnits, err := quadlet.List(realContainersPath)
	if err != nil {
		return err
	}
	index := make(map[string]string, len(allUnits)) // baseName -> path
	for _, u := range allUnits {
		index[u.BaseName()] = u.Path
	}

	var anyChanged bool
	var failures int

	for _, unitName := range unitNames {
		quadletFile, ok := index[unitName]
		if !ok {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error: quadlet file for unit %q not found", unitName)))
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

		newContent, changed := apply(string(content))
		if !changed {
			fmt.Println(shared.WarningStyle.Render(unchangedMessage(unitName)))
			continue
		}

		if err := writeQuadletFile(quadletFile, []byte(newContent)); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error writing file: %v", err)))
			failures++
			continue
		}

		fmt.Println(shared.CheckMark + " " + fmt.Sprintf(changedMessage, shared.FilePathStyle.Render(quadletFile)))
		fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully %sd %s", actionName, unitName)))
		anyChanged = true
	}

	if anyChanged {
		fmt.Println("  Running systemctl --user daemon-reload...")
		if _, err := systemd.DaemonReload(); err != nil {
			return cmdutil.Wrap(err, "running daemon-reload")
		}
		fmt.Println(shared.CheckMark + " Daemon reloaded.")
	}

	if failures > 0 {
		return cmdutil.Errorf("%d unit(s) failed to %s", failures, actionName)
	}

	fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Successfully processed %d unit(s)", len(unitNames))))
	return nil
}
