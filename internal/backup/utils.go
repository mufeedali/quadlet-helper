package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/systemd"
)

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

func GetServiceFilePath(backupName string) (string, error) {
	userDir, err := systemd.UserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, BackupServiceName(backupName)), nil
}

func GetTimerFilePath(backupName string) (string, error) {
	userDir, err := systemd.UserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, BackupTimerName(backupName)), nil
}

func GetNotificationServiceFilePath(backupName string) (string, error) {
	userDir, err := systemd.UserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, fmt.Sprintf("backup-notify@%s.service", backupName)), nil
}

func BaseEnv(config *Config) []string {
	return append(os.Environ(), config.Environment...)
}

func ResticEnv(config *Config) []string {
	env := BaseEnv(config)
	if config.Options.PasswordFile != "" {
		env = append(env, fmt.Sprintf("RESTIC_PASSWORD_FILE=%s", config.Options.PasswordFile))
	}
	return append(env, fmt.Sprintf("RESTIC_REPOSITORY=%s", config.Destination.Repository))
}

func RcloneDestPath(remote, source string, totalSources int) string {
	if totalSources > 1 {
		return filepath.Join(remote, filepath.Base(source))
	}
	return remote
}

func RsyncDestPath(destinationPath, source string, totalSources int) string {
	trimmedSource := strings.TrimRight(source, string(os.PathSeparator))
	if trimmedSource == "" {
		trimmedSource = source
	}

	if totalSources > 1 {
		return filepath.Join(destinationPath, filepath.Base(trimmedSource))
	}

	if strings.HasSuffix(source, string(os.PathSeparator)) {
		return destinationPath
	}

	if info, err := os.Stat(destinationPath); err == nil && !info.IsDir() {
		return destinationPath
	}

	return filepath.Join(destinationPath, filepath.Base(trimmedSource))
}

func RsyncArgs(config *Config, dryRun bool) []string {
	args := []string{}
	if dryRun {
		args = append(args, "--dry-run")
	}
	if config.Options.Archive {
		args = append(args, "-a")
	}
	if config.Options.Compress {
		args = append(args, "-z")
	}
	if config.Options.Delete {
		args = append(args, "--delete")
	}
	args = append(args, "-v", "--progress")
	for _, exclude := range config.Options.Exclude {
		args = append(args, "--exclude", exclude)
	}
	args = append(args, config.Source...)
	return append(args, config.Destination.Path)
}

func ResticBackupArgs(config *Config) []string {
	args := []string{"backup"}
	for _, exclude := range config.Options.Exclude {
		args = append(args, "--exclude", exclude)
	}
	return append(args, config.Source...)
}

func RcloneBaseArgs(config *Config, dryRun bool) []string {
	args := []string{"sync"}
	if dryRun {
		args = append(args, "--dry-run")
	}
	if config.Options.Transfers > 0 {
		args = append(args, "--transfers", fmt.Sprintf("%d", config.Options.Transfers))
	}
	if config.Options.Checkers > 0 {
		args = append(args, "--checkers", fmt.Sprintf("%d", config.Options.Checkers))
	}
	if config.Options.BandwidthLimit != "" {
		args = append(args, "--bwlimit", config.Options.BandwidthLimit)
	}
	for _, exclude := range config.Options.Exclude {
		args = append(args, "--exclude", exclude)
	}
	return append(args, "-v", "--progress")
}
