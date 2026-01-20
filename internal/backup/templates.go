package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetServiceTemplate returns the systemd service template for a backup
func GetServiceTemplate(executablePath, configPath, backupName string, config *Config) string {
	var template strings.Builder
	template.WriteString(`[Unit]
Description=Backup: %s
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
ExecStart=%s backup run %s`)

	// Prepend user's .local/bin to PATH for tools like restic, rclone installed locally
	template.WriteString("\nEnvironment=PATH=%%h/.local/bin:%%h/.local/share/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin")

	// Add environment variables
	if len(config.Environment) > 0 {
		for _, env := range config.Environment {
			template.WriteString(fmt.Sprintf("\nEnvironment=%s", env))
		}
	}

	template.WriteString(`
StandardOutput=journal
StandardError=journal`)

	// Add verification step if enabled
	if config.Verification.Enabled && config.Verification.AutoVerify {
		template.WriteString(fmt.Sprintf("\nExecStartPost=%s backup verify %s", executablePath, backupName))
	}

	// Add cleanup step if retention is configured
	if config.Retention.KeepDays > 0 || config.Retention.KeepDaily > 0 {
		template.WriteString(fmt.Sprintf("\nExecStopPost=%s backup cleanup %s", executablePath, backupName))
	}

	return fmt.Sprintf(template.String(), backupName, executablePath, backupName)
}

// GetTimerTemplate returns the systemd timer template for a backup
func GetTimerTemplate(backupName, schedule string) (string, error) {
	onCalendar, err := ParseSchedule(schedule)
	if err != nil {
		return "", err
	}

	template := `[Unit]
Description=Backup timer for %s

[Timer]
OnCalendar=%s
Persistent=true
Unit=%s-backup.service

[Install]
WantedBy=timers.target
`

	return fmt.Sprintf(template, backupName, onCalendar, backupName), nil
}

// GetNotificationServiceTemplate returns the systemd notification service template
func GetNotificationServiceTemplate(executablePath, backupName string) string {
	template := `[Unit]
Description=Email notification for failed backup: %s

[Service]
Type=oneshot
ExecStart=%s backup notify %s failure
StandardOutput=journal
StandardError=journal
`

	return fmt.Sprintf(template, backupName, executablePath, backupName)
}

// GetSystemdUserDir returns the systemd user directory path
func GetSystemdUserDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error finding home directory: %w", err)
	}
	return filepath.Join(home, ".config", "systemd", "user"), nil
}

// GetServiceFilePath returns the path to the service file
func GetServiceFilePath(backupName string) (string, error) {
	systemdDir, err := GetSystemdUserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(systemdDir, fmt.Sprintf("%s-backup.service", backupName)), nil
}

// GetTimerFilePath returns the path to the timer file
func GetTimerFilePath(backupName string) (string, error) {
	systemdDir, err := GetSystemdUserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(systemdDir, fmt.Sprintf("%s-backup.timer", backupName)), nil
}

// GetNotificationServiceFilePath returns the path to the notification service file
func GetNotificationServiceFilePath(backupName string) (string, error) {
	systemdDir, err := GetSystemdUserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(systemdDir, fmt.Sprintf("backup-notify@%s.service", backupName)), nil
}
