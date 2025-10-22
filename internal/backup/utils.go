package backup

import "fmt"

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

// BackupTimerName returns the systemd timer name for a backup.
func BackupTimerName(backupName string) string {
	return fmt.Sprintf("%s-backup.timer", backupName)
}

// BackupServiceName returns the systemd service name for a backup.
func BackupServiceName(backupName string) string {
	return fmt.Sprintf("%s-backup.service", backupName)
}
