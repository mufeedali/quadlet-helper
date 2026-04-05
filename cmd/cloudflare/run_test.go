package cloudflare

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateCloudflareIPsInConfigUpdatesDecodedYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "traefik.yaml")
	content := []byte("cloudflare-ips:\n  trustedIPs:\n    - 1.1.1.1/32\n")

	if err := os.WriteFile(configPath, content, 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	config, err := readTraefikConfig(configPath)
	if err != nil {
		t.Fatalf("readTraefikConfig() error = %v", err)
	}

	changed, updated, err := updateCloudflareIPsInConfig(config, []string{"2.2.2.2/32", "3.3.3.3/32"})
	if err != nil {
		t.Fatalf("updateCloudflareIPsInConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("updateCloudflareIPsInConfig() changed = false, want true")
	}

	cfSection, ok := updated["cloudflare-ips"].(map[string]any)
	if !ok {
		t.Fatalf("updated cloudflare-ips type = %T, want map[string]any", updated["cloudflare-ips"])
	}

	trustedIPs, ok := cfSection["trustedIPs"].([]string)
	if !ok {
		t.Fatalf("trustedIPs type = %T, want []string", cfSection["trustedIPs"])
	}

	if len(trustedIPs) != 2 || trustedIPs[0] != "2.2.2.2/32" || trustedIPs[1] != "3.3.3.3/32" {
		t.Fatalf("trustedIPs = %v, want replacement values", trustedIPs)
	}
}

func TestUpdateCloudflareIPsInConfigReturnsErrorForMissingSection(t *testing.T) {
	_, _, err := updateCloudflareIPsInConfig(map[string]any{}, []string{"1.1.1.1/32"})
	if err == nil {
		t.Fatal("updateCloudflareIPsInConfig() error = nil, want non-nil")
	}
}
