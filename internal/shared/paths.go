package shared

import "path/filepath"

func TraefikConfigPath(containersDir string) string {
	return filepath.Join(containersDir, "traefik", "container-config", "traefik", "traefik.yaml")
}

func TraefikExamplePath(containersDir string) string {
	return filepath.Join(containersDir, "traefik", "traefik.yaml.example")
}
