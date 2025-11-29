package unit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/viper"
)

// resolveServiceNames resolves a list of unit names to their corresponding systemd service names.
func resolveServiceNames(unitNames []string) ([]string, error) {
	containersPath := viper.GetString("containers-path")
	realContainersPath := shared.ResolveContainersDir(containersPath)

	var services []string
	for _, unitName := range unitNames {
		serviceName, err := resolveServiceName(realContainersPath, unitName)
		if err != nil {
			return nil, err
		}
		services = append(services, serviceName)
	}
	return services, nil
}

// resolveServiceName resolves a single unit name to its systemd service name.
func resolveServiceName(dir, unitName string) (string, error) {
	// First, try to find the file to determine the extension
	quadletFile, err := findQuadletFile(dir, unitName)
	if err != nil {
		return "", err
	}

	ext := filepath.Ext(quadletFile)
	return getServiceNameFromExtension(unitName, ext), nil
}

// getServiceNameFromExtension returns the systemd service name based on the unit name and extension.
func getServiceNameFromExtension(unitName, ext string) string {
	switch ext {
	case ".pod":
		return unitName + "-pod.service"
	case ".volume":
		return unitName + "-volume.service"
	case ".network":
		return unitName + "-network.service"
	case ".image":
		return unitName + "-image.service"
	case ".build":
		return unitName + "-build.service"
	case ".artifact":
		return unitName + "-artifact.service"
	default:
		// .container, .kube, and others default to .service
		return unitName + ".service"
	}
}

// findQuadletFile searches for a quadlet file with the given unit name in the directory.
func findQuadletFile(dir, unitName string) (string, error) {
	var foundPath string
	err := shared.WalkWithSymlinks(dir, func(path string, info os.FileInfo) error {
		if !info.IsDir() {
			baseName := filepath.Base(path)
			ext := filepath.Ext(baseName)
			// Check if the file matches the unit name (ignoring extension)
			if len(baseName) > len(ext) && baseName[:len(baseName)-len(ext)] == unitName {
				// Verify it's a valid quadlet extension
				if isQuadletUnit(ext) {
					foundPath = path
					return filepath.SkipDir // Stop searching
				}
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

// isQuadletUnit checks if the file extension corresponds to a valid Quadlet unit.
func isQuadletUnit(ext string) bool {
	validExtensions := map[string]bool{
		".container": true,
		".pod":       true,
		".network":   true,
		".volume":    true,
		".kube":      true,
		".image":     true,
		".build":     true,
		".artifact":  true, // Added based on docs
	}
	return validExtensions[ext]
}
