package backup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// GetConfigDir returns the backup configuration directory
func GetConfigDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error finding home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")
	}
	configDir := filepath.Join(configHome, "quadlet-helper", "backups")
	return configDir, nil
}

// GetConfigPath returns the full path to a backup config file
func GetConfigPath(name string) (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, name+".yaml"), nil
}

// LoadConfig loads a backup configuration from file
func LoadConfig(name string) (*Config, error) {
	configPath, err := GetConfigPath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&config); err != nil {
		return nil, fmt.Errorf("error parsing config file:\n%w", err)
	}
	config = config.Normalized()

	return &config, nil
}

// SaveConfig saves a backup configuration to file
func SaveConfig(config *Config) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	configPath, err := GetConfigPath(config.Name)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(config); err != nil {
		_ = enc.Close()
		return fmt.Errorf("error marshaling config: %w", err)
	}
	_ = enc.Close()
	data := buf.Bytes()

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// ListConfigs returns all backup configuration names
func ListConfigs() ([]string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, fmt.Errorf("error reading config directory: %w", err)
	}

	var configs []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".yaml")
		configs = append(configs, name)
	}

	return configs, nil
}
