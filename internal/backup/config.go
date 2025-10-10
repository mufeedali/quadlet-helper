package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

// BackupType represents the type of backup tool
type BackupType string

const (
	BackupTypeRsync  BackupType = "rsync"
	BackupTypeRestic BackupType = "restic"
	BackupTypeRclone BackupType = "rclone"
)

// Config represents a backup configuration
type Config struct {
	Name          string        `yaml:"name"`
	Type          BackupType    `yaml:"type"`
	Schedule      string        `yaml:"schedule"`
	Source        []string      `yaml:"source"`
	Destination   Destination   `yaml:"destination"`
	Options       Options       `yaml:"options,omitempty"`
	Verification  Verification  `yaml:"verification,omitempty"`
	Retention     Retention     `yaml:"retention,omitempty"`
	Notifications Notifications `yaml:"notifications,omitempty"`
	Hooks         Hooks         `yaml:"hooks,omitempty"`
	Environment   []string      `yaml:"environment,omitempty"`
}

// Destination varies by backup type
type Destination struct {
	Remote     string `yaml:"remote,omitempty"`     // For rclone
	Path       string `yaml:"path,omitempty"`       // For rsync
	Repository string `yaml:"repository,omitempty"` // For restic
}

// Options contains backup-specific options
type Options struct {
	// Rclone options
	Transfers      int      `yaml:"transfers,omitempty"`
	Checkers       int      `yaml:"checkers,omitempty"`
	BandwidthLimit string   `yaml:"bandwidth_limit,omitempty"`
	Exclude        []string `yaml:"exclude,omitempty"`

	// Rsync options
	Archive  bool `yaml:"archive,omitempty"`
	Compress bool `yaml:"compress,omitempty"`
	Delete   bool `yaml:"delete,omitempty"`

	// Restic options
	PasswordFile string `yaml:"password_file,omitempty"`
	KeepDaily    int    `yaml:"keep_daily,omitempty"`
	KeepWeekly   int    `yaml:"keep_weekly,omitempty"`
}

// Verification settings
type Verification struct {
	Enabled    bool   `yaml:"enabled"`
	AutoVerify bool   `yaml:"auto_verify"`
	Method     string `yaml:"method,omitempty"` // check, checksum, size, cryptcheck
}

// Retention settings
type Retention struct {
	KeepDays    int `yaml:"keep_days,omitempty"`
	KeepDaily   int `yaml:"keep_daily,omitempty"`
	KeepWeekly  int `yaml:"keep_weekly,omitempty"`
	KeepMonthly int `yaml:"keep_monthly,omitempty"`
}

// Notifications settings
type Notifications struct {
	Enabled   bool        `yaml:"enabled"`
	OnFailure bool        `yaml:"on_failure"`
	OnSuccess bool        `yaml:"on_success"`
	Email     EmailConfig `yaml:"email,omitempty"`
}

// EmailConfig for email notifications
type EmailConfig struct {
	To   string     `yaml:"to"`
	From string     `yaml:"from"`
	SMTP SMTPConfig `yaml:"smtp,omitempty"`
}

// SMTPConfig for SMTP settings
type SMTPConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	PasswordFile string `yaml:"password_file"`
	TLS          bool   `yaml:"tls"`
}

// Hooks for pre/post backup scripts
type Hooks struct {
	PreBackup  string `yaml:"pre_backup,omitempty"`
	PostBackup string `yaml:"post_backup,omitempty"`
	OnFailure  string `yaml:"on_failure,omitempty"`
}

// GetConfigDir returns the backup configuration directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error finding home directory: %w", err)
	}
	configDir := filepath.Join(home, ".config", "quadlet-helper", "backups")
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
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

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

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
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
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			name := strings.TrimSuffix(entry.Name(), ".yaml")
			configs = append(configs, name)
		}
	}

	return configs, nil
}

// Validate validates the backup configuration
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("backup name is required")
	}

	if c.Type != BackupTypeRsync && c.Type != BackupTypeRestic && c.Type != BackupTypeRclone {
		return fmt.Errorf("invalid backup type: %s (must be rsync, restic, or rclone)", c.Type)
	}

	if len(c.Source) == 0 {
		return fmt.Errorf("at least one source path is required")
	}

	// Validate destination based on type
	switch c.Type {
	case BackupTypeRclone:
		if c.Destination.Remote == "" {
			return fmt.Errorf("destination.remote is required for rclone backups")
		}
	case BackupTypeRsync:
		if c.Destination.Path == "" {
			return fmt.Errorf("destination.path is required for rsync backups")
		}
	case BackupTypeRestic:
		if c.Destination.Repository == "" {
			return fmt.Errorf("destination.repository is required for restic backups")
		}
	}

	// Validate schedule format
	if c.Schedule == "" {
		return fmt.Errorf("schedule is required")
	}

	// Validate email config if notifications enabled
	if c.Notifications.Enabled {
		if c.Notifications.Email.To == "" {
			return fmt.Errorf("email.to is required when notifications are enabled")
		}
		if c.Notifications.Email.SMTP.Host == "" {
			return fmt.Errorf("email.smtp.host is required when notifications are enabled")
		}
		if c.Notifications.Email.SMTP.Port == 0 {
			return fmt.Errorf("email.smtp.port is required when notifications are enabled")
		}
	}

	return nil
}

// GetDestination returns the destination string for the backup type
func (c *Config) GetDestination() string {
	switch c.Type {
	case BackupTypeRclone:
		return c.Destination.Remote
	case BackupTypeRsync:
		return c.Destination.Path
	case BackupTypeRestic:
		return c.Destination.Repository
	default:
		return ""
	}
}

// ParseSchedule converts schedule string to systemd OnCalendar format
func ParseSchedule(schedule string) (string, error) {
	schedule = strings.TrimSpace(strings.ToLower(schedule))

	// Common schedule patterns
	schedules := map[string]string{
		"hourly":           "*-*-* *:00:00",
		"daily":            "*-*-* 02:00:00",
		"daily 02:00":      "*-*-* 02:00:00",
		"weekly":           "Mon *-*-* 02:00:00",
		"weekly sun 03:00": "Sun *-*-* 03:00:00",
		"monthly":          "*-*-01 02:00:00",
	}

	// Check if it's a predefined pattern
	if cal, ok := schedules[schedule]; ok {
		return cal, nil
	}

	// Parse "daily HH:MM" format
	if strings.HasPrefix(schedule, "daily ") {
		time := strings.TrimPrefix(schedule, "daily ")
		return fmt.Sprintf("*-*-* %s:00", time), nil
	}

	// Parse "weekly DAY HH:MM" format
	if strings.HasPrefix(schedule, "weekly ") {
		parts := strings.Fields(strings.TrimPrefix(schedule, "weekly "))
		if len(parts) == 2 {
			caser := cases.Title(language.English)
			day := caser.String(parts[0])
			return fmt.Sprintf("%s *-*-* %s:00", day, parts[1]), nil
		}
	}

	// If it looks like a systemd calendar format, use it directly
	if strings.Contains(schedule, "*") || strings.Contains(schedule, ":") {
		return schedule, nil
	}

	return "", fmt.Errorf("invalid schedule format: %s", schedule)
}
