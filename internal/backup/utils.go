package backup

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
