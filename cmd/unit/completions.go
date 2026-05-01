package unit

import (
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/quadlet"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// allQuadletTypes lists every type that podman quadlet understands.
var allQuadletTypes = []string{"artifact", "build", "container", "image", "kube", "network", "pod", "volume"}

// typeCompletionFunc completes values for the --type flag.
func typeCompletionFunc(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return allQuadletTypes, cobra.ShellCompDirectiveNoFileComp
}

func typeFilter(cmd *cobra.Command) map[string]bool {
	types, err := cmd.Flags().GetStringSlice("type")
	if err != nil || len(types) == 0 {
		return nil
	}
	m := make(map[string]bool, len(types))
	for _, t := range types {
		m[t] = true
	}
	return m
}

// unitCompletionFunc provides dynamic completion for quadlet unit names.
// Respects --type if the command has that flag.
func unitCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	containersPath := viper.GetString("containers-path")
	realContainersPath := shared.ResolveContainersDir(containersPath)

	units, err := quadlet.List(realContainersPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	filter := typeFilter(cmd)
	var completions []string
	for _, u := range units {
		if filter != nil && !filter[u.UnitType()] {
			continue
		}
		name := u.BaseName()
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// activeUnitCompletionFunc provides dynamic completion for active quadlet unit names.
// Respects --type if the command has that flag.
func activeUnitCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	containersPath := viper.GetString("containers-path")
	realContainersPath := shared.ResolveContainersDir(containersPath)

	units, err := quadlet.List(realContainersPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	filter := typeFilter(cmd)
	var completions []string
	for _, u := range units {
		if !u.IsActive() {
			continue
		}
		if filter != nil && !filter[u.UnitType()] {
			continue
		}
		name := u.BaseName()
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
