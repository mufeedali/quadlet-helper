package unit

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// unitCompletionFunc provides dynamic completion for quadlet unit names.
func unitCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	containersPath := viper.GetString("containers-path")
	realContainersPath := shared.ResolveContainersDir(containersPath)

	var completions []string

	err := shared.WalkWithSymlinks(realContainersPath, func(path string, info os.FileInfo) error {
		if !info.IsDir() {
			ext := filepath.Ext(info.Name())
			if isQuadletUnit(ext) {
				unitName := strings.TrimSuffix(info.Name(), ext)
				// Add to completions if it matches the currently typed prefix
				if strings.HasPrefix(unitName, toComplete) {
					completions = append(completions, unitName)
				}
			}
		}
		return nil
	})

	if err != nil {
		// In case of error, return no completions
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
