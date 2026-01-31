package unit

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// unitCompletionFunc provides dynamic completion for quadlet unit names.
func unitCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	containersPath := viper.GetString("containers-path")
	realContainersPath := shared.ResolveContainersDir(containersPath)

	var completions []string

	err := shared.WalkWithSymlinks(realContainersPath, func(path string, d fs.DirEntry) error {
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(d.Name())
		if !isQuadletUnit(ext) {
			return nil
		}

		unitName := strings.TrimSuffix(d.Name(), ext)
		// Add to completions if it matches the currently typed prefix
		if strings.HasPrefix(unitName, toComplete) {
			completions = append(completions, unitName)
		}
		return nil
	})

	if err != nil {
		// In case of error, return no completions
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// activeUnitCompletionFunc provides dynamic completion for active quadlet unit names.
func activeUnitCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	containersPath := viper.GetString("containers-path")
	realContainersPath := shared.ResolveContainersDir(containersPath)

	activeServices, err := systemd.ListActiveServices()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Create a map for faster lookup
	activeMap := make(map[string]bool)
	for _, s := range activeServices {
		activeMap[s] = true
	}

	var completions []string

	err = shared.WalkWithSymlinks(realContainersPath, func(path string, d fs.DirEntry) error {
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(d.Name())
		if !isQuadletUnit(ext) {
			return nil
		}

		unitName := strings.TrimSuffix(d.Name(), ext)
		serviceName := getServiceNameFromExtension(unitName, ext)

		// Add to completions if it matches the currently typed prefix AND is active
		if activeMap[serviceName] && strings.HasPrefix(unitName, toComplete) {
			completions = append(completions, unitName)
		}

		return nil
	})

	if err != nil {
		// In case of error, return no completions
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
