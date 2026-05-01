package unit

import (
	"fmt"
	"os"

	"github.com/mufeedali/quadlet-helper/internal/quadlet"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/viper"
)

// resolveServiceNames resolves a list of unit names to their systemd service names
// via a single podman quadlet list call. Only units matching the given types are considered.
func resolveServiceNames(unitNames []string, types []string) ([]string, error) {
	containersPath := viper.GetString("containers-path")
	realContainersPath := shared.ResolveContainersDir(containersPath)

	allUnits, err := quadlet.List(realContainersPath)
	if err != nil {
		return nil, err
	}

	typeSet := make(map[string]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}

	index := make(map[string]*quadlet.Unit, len(allUnits))
	for i, u := range allUnits {
		if len(typeSet) == 0 || typeSet[u.UnitType()] {
			name := u.BaseName()
			index[name] = &allUnits[i]
		}
	}

	services := make([]string, 0, len(unitNames))
	for _, unitName := range unitNames {
		u, ok := index[unitName]
		if !ok {
			return nil, fmt.Errorf("quadlet unit %q not found", unitName)
		}
		services = append(services, u.UnitName)
	}
	return services, nil
}

func writeQuadletFile(path string, content []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	return os.WriteFile(path, content, info.Mode().Perm())
}
