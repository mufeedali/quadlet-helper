package backup

import (
	"fmt"
)

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

	return nil
}
