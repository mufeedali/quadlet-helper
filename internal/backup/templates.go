package backup

import (
	"fmt"
	"strings"
)

// GetServiceTemplate returns the systemd service template for a backup
func GetServiceTemplate(executablePath, backupName string, config *Config) string {
	safeBackupName := sanitizeUnitLine(backupName)
	var template strings.Builder
	fmt.Fprintf(&template, `[Unit]
Description=Backup: %s
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
ExecStart=%q backup run %q`, safeBackupName, executablePath, backupName)

	// Prepend user's .local/bin to PATH for tools like restic, rclone installed locally
	template.WriteString("\nEnvironment=PATH=%%h/.local/bin:%%h/.local/share/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin")

	// Add environment variables
	if len(config.Environment) > 0 {
		for _, env := range config.Environment {
			fmt.Fprintf(&template, "\nEnvironment=%q", env)
		}
	}

	template.WriteString(`
StandardOutput=journal
StandardError=journal`)

	// Add verification step if enabled
	if config.Verification.Enabled && config.Verification.AutoVerify {
		fmt.Fprintf(&template, "\nExecStartPost=%q backup verify %q", executablePath, backupName)
	}

	// Add cleanup step if retention is configured
	if config.Retention.KeepDays > 0 || config.Retention.KeepDaily > 0 {
		fmt.Fprintf(&template, "\nExecStopPost=%q backup cleanup %q", executablePath, backupName)
	}

	return template.String()
}

// GetTimerTemplate returns the systemd timer template for a backup
func GetTimerTemplate(backupName, schedule string) (string, error) {
	onCalendar, err := ParseSchedule(schedule)
	if err != nil {
		return "", err
	}
	safeBackupName := sanitizeUnitLine(backupName)

	template := `[Unit]
Description=Backup timer for %s

[Timer]
OnCalendar=%s
Persistent=true
Unit=%s-backup.service

[Install]
WantedBy=timers.target
`

	return fmt.Sprintf(template, safeBackupName, onCalendar, safeBackupName), nil
}

// GetNotificationServiceTemplate returns the systemd notification service template
func GetNotificationServiceTemplate(executablePath, backupName string) string {
	safeBackupName := sanitizeUnitLine(backupName)
	template := `[Unit]
Description=Email notification for failed backup: %s

[Service]
Type=oneshot
ExecStart=%q backup notify %q failure
StandardOutput=journal
StandardError=journal
`

	return fmt.Sprintf(template, safeBackupName, executablePath, backupName)
}

func sanitizeUnitLine(value string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(value)
}
