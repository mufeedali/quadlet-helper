package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDoesNotMutateVerificationMethod(t *testing.T) {
	config := Config{
		Name:     "demo",
		Type:     BackupTypeRsync,
		Schedule: "daily",
		Source:   []string{"/tmp/source"},
		Destination: Destination{
			Path: "/tmp/dest",
		},
		Verification: Verification{
			Enabled: true,
		},
	}

	if err := config.Validate(); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
	if config.Verification.Method != "" {
		t.Fatalf("Validate() mutated verification method to %q", config.Verification.Method)
	}
}

func TestLoadConfigNormalizesVerificationMethod(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	backupDir := filepath.Join(configHome, "quadlet-helper", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	configPath := filepath.Join(backupDir, "demo.yaml")
	configData := []byte(`name: demo
type: rsync
schedule: daily
source:
  - /tmp/source
destination:
  path: /tmp/dest
verification:
  enabled: true
`)
	if err := os.WriteFile(configPath, configData, 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	config, err := LoadConfig("demo")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.Verification.Method != VerificationMethodSize {
		t.Fatalf("Verification.Method = %q, want %q", config.Verification.Method, VerificationMethodSize)
	}
}
