package quadlet

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Unit represents a quadlet unit as reported by "podman quadlet list --format json".
type Unit struct {
	Name     string `json:"Name"`     // quadlet file name, e.g. "actual.container"
	UnitName string `json:"UnitName"` // systemd service name, e.g. "actual.service"
	Path     string `json:"Path"`     // absolute path to the quadlet file
	Status   string `json:"Status"`   // e.g. "active/running", "active/exited", "inactive/dead"
	App      string `json:"App"`
}

// IsActive reports whether the unit's active state is "active".
func (u Unit) IsActive() bool {
	return strings.HasPrefix(u.Status, "active/")
}

// UnitType returns the quadlet file type derived from the Name extension,
// e.g. "container", "network", "volume".
func (u Unit) UnitType() string {
	if idx := strings.LastIndex(u.Name, "."); idx >= 0 {
		return u.Name[idx+1:]
	}
	return ""
}

// BaseName returns the unit's name without the file extension.
func (u Unit) BaseName() string {
	if idx := strings.LastIndex(u.Name, "."); idx >= 0 {
		return u.Name[:idx]
	}
	return u.Name
}

// List returns all quadlet units reported by "podman quadlet list --format json".
// If pathPrefix is non-empty, only units whose Path starts with that prefix are returned.
func List(pathPrefix string) ([]Unit, error) {
	cmd := exec.Command("podman", "quadlet", "list", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running podman quadlet list: %w", err)
	}

	var units []Unit
	if err := json.Unmarshal(output, &units); err != nil {
		return nil, fmt.Errorf("parsing podman quadlet list output: %w", err)
	}

	if pathPrefix == "" {
		return units, nil
	}

	prefix := strings.TrimRight(pathPrefix, "/")
	var filtered []Unit
	for _, u := range units {
		if strings.HasPrefix(u.Path, prefix+"/") || u.Path == prefix {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}
