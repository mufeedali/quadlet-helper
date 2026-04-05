package backup

import (
	"fmt"
	"regexp"
)

var validBackupName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_\-.]*$`)

// Validate validates the backup configuration
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("backup name is required")
	}
	if !validBackupName.MatchString(c.Name) {
		return fmt.Errorf("backup name %q contains invalid characters (only alphanumerics, hyphens, underscores, and dots are allowed)", c.Name)
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

	method := c.Verification.Method
	if c.Verification.Enabled || method != "" {
		switch c.Type {
		case BackupTypeRsync:
			if method == "" {
				method = VerificationMethodSize
			}
			switch method {
			case VerificationMethodSize, VerificationMethodChecksum:
				// ok
			default:
				return fmt.Errorf("verification method %q not supported for rsync", method)
			}

		case BackupTypeRestic:
			if method == "" {
				method = VerificationMethodCheck
			}
			if method != VerificationMethodCheck {
				return fmt.Errorf("verification method %q not supported for restic", method)
			}

		case BackupTypeRclone:
			if method == "" {
				method = VerificationMethodCheck
			}
			switch method {
			case VerificationMethodCheck, VerificationMethodCryptCheck, VerificationMethodSize:
				// ok
			default:
				return fmt.Errorf("verification method %q not supported for rclone", method)
			}
		}
	}

	return nil
}

func (c Config) Normalized() Config {
	normalized := c
	if !normalized.Verification.Enabled && normalized.Verification.Method == "" {
		return normalized
	}

	switch normalized.Type {
	case BackupTypeRsync:
		if normalized.Verification.Method == "" {
			normalized.Verification.Method = VerificationMethodSize
		}
	case BackupTypeRestic:
		if normalized.Verification.Method == "" {
			normalized.Verification.Method = VerificationMethodCheck
		}
	case BackupTypeRclone:
		if normalized.Verification.Method == "" {
			normalized.Verification.Method = VerificationMethodCheck
		}
	}

	return normalized
}
