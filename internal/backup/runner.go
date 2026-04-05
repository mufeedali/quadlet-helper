package backup

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"
)

// RunResult represents the result of a backup run
type RunResult struct {
	Success   bool
	StartTime time.Time
	EndTime   time.Time
	Output    string
	Error     error
}

// CheckToolAvailable checks if a backup tool is installed and available in PATH
func CheckToolAvailable(backupType BackupType) (bool, error) {
	var toolName string
	switch backupType {
	case BackupTypeRsync:
		toolName = "rsync"
	case BackupTypeRestic:
		toolName = "restic"
	case BackupTypeRclone:
		toolName = "rclone"
	default:
		return false, fmt.Errorf("unknown backup type: %s", backupType)
	}

	_, err := exec.LookPath(toolName)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetInstallInstructions returns installation instructions for a backup tool
func GetInstallInstructions(backupType BackupType) string {
	switch backupType {
	case BackupTypeRsync:
		return `rsync is not installed. Install it with:
  - Ubuntu/Debian: sudo apt install rsync
  - Fedora/RHEL: sudo dnf install rsync
  - macOS: brew install rsync (usually pre-installed)
  - Arch: sudo pacman -S rsync`
	case BackupTypeRestic:
		return `restic is not installed. Install it with:
  - Ubuntu/Debian: sudo apt install restic
  - Fedora/RHEL: sudo dnf install restic
  - macOS: brew install restic
  - Arch: sudo pacman -S restic
  - Or download from: https://restic.net/`
	case BackupTypeRclone:
		return `rclone is not installed. Install it with:
  - Ubuntu/Debian: sudo apt install rclone
  - Fedora/RHEL: sudo dnf install rclone
  - macOS: brew install rclone
  - Arch: sudo pacman -S rclone
  - Or: curl https://rclone.org/install.sh | sudo bash`
	default:
		return fmt.Sprintf("Unknown backup type: %s", backupType)
	}
}

// Run executes a backup based on its configuration
func Run(config *Config, dryRun bool) (*RunResult, error) {
	normalized := config.Normalized()
	config = &normalized

	result := &RunResult{
		StartTime: time.Now(),
	}

	// Check if the required tool is available
	available, checkErr := CheckToolAvailable(config.Type)
	if checkErr != nil {
		result.Error = checkErr
		result.EndTime = time.Now()
		return result, checkErr
	}
	if !available {
		result.Error = fmt.Errorf("%s is not installed or not in PATH\n\n%s",
			config.Type, GetInstallInstructions(config.Type))
		result.EndTime = time.Now()
		return result, result.Error
	}

	// Run pre-backup hook if configured
	if config.Hooks.PreBackup != "" {
		if err := runHook(config.Hooks.PreBackup); err != nil {
			result.Error = fmt.Errorf("pre-backup hook failed: %w", err)
			result.EndTime = time.Now()
			return result, result.Error
		}
	}

	// Execute backup based on type
	var output string
	var err error

	switch config.Type {
	case BackupTypeRsync:
		output, err = runRsyncBackup(config, dryRun)
	case BackupTypeRestic:
		output, err = runResticBackup(config, dryRun)
	case BackupTypeRclone:
		output, err = runRcloneBackup(config, dryRun)
	default:
		err = fmt.Errorf("unsupported backup type: %s", config.Type)
	}

	result.Output = output
	result.EndTime = time.Now()

	if err != nil {
		result.Error = err
		result.Success = false

		// Run failure hook if configured
		if config.Hooks.OnFailure != "" {
			_ = runHook(config.Hooks.OnFailure)
		}

		return result, err
	}

	result.Success = true

	// Run post-backup hook if configured
	if config.Hooks.PostBackup != "" {
		if err := runHook(config.Hooks.PostBackup); err != nil {
			result.Error = fmt.Errorf("post-backup hook failed: %w", err)
			result.Success = false
			return result, result.Error
		}
	}

	return result, nil
}

// runRsyncBackup executes an rsync backup
func runRsyncBackup(config *Config, dryRun bool) (string, error) {
	output, err := runCommandStreaming("rsync", RsyncArgs(config, dryRun), BaseEnv(config))
	if err != nil {
		return output, fmt.Errorf("rsync failed: %w", err)
	}

	return output, nil
}

// runResticBackup executes a restic backup
func runResticBackup(config *Config, dryRun bool) (string, error) {
	env := ResticEnv(config)

	if dryRun {
		return runCommandStreaming("restic", []string{"snapshots", "--latest", "1"}, env)
	}

	output, err := runCommandStreaming("restic", ResticBackupArgs(config), env)
	if err != nil {
		return output, fmt.Errorf("restic backup failed: %w", err)
	}

	return output, nil
}

// runRcloneBackup executes an rclone backup
func runRcloneBackup(config *Config, dryRun bool) (string, error) {
	args := RcloneBaseArgs(config, dryRun)
	env := BaseEnv(config)

	if len(config.Source) == 1 {
		output, err := runCommandStreaming("rclone", append(args, config.Source[0], config.Destination.Remote), env)
		if err != nil {
			return output, fmt.Errorf("rclone failed: %w", err)
		}
		return output, nil
	} else {
		var allOutput strings.Builder
		for _, source := range config.Source {
			srcArgs := slices.Clone(args)
			srcArgs = append(srcArgs, source, RcloneDestPath(config.Destination.Remote, source, len(config.Source)))

			output, err := runCommandStreaming("rclone", srcArgs, env)
			allOutput.WriteString(output)
			if !strings.HasSuffix(output, "\n") {
				allOutput.WriteString("\n")
			}

			if err != nil {
				return allOutput.String(), fmt.Errorf("rclone failed for %s: %w", source, err)
			}
		}
		return allOutput.String(), nil
	}
}

// runHook executes a hook script
func runHook(hookPath string) error {
	cmd := exec.Command("/bin/sh", "-c", hookPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hook failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// Cleanup performs retention cleanup based on backup type
func Cleanup(config *Config) error {
	normalized := config.Normalized()
	config = &normalized

	switch config.Type {
	case BackupTypeRestic:
		return cleanupRestic(config)
	case BackupTypeRclone:
		return cleanupRclone(config)
	case BackupTypeRsync:
		// Rsync doesn't have built-in retention, would need custom logic
		return nil
	default:
		return nil
	}
}

// cleanupRestic performs restic retention cleanup
func cleanupRestic(config *Config) error {
	if config.Options.KeepDaily == 0 && config.Options.KeepWeekly == 0 {
		return nil
	}

	env := ResticEnv(config)

	args := []string{"forget", "--prune"}

	if config.Options.KeepDaily > 0 {
		args = append(args, "--keep-daily", fmt.Sprintf("%d", config.Options.KeepDaily))
	}
	if config.Options.KeepWeekly > 0 {
		args = append(args, "--keep-weekly", fmt.Sprintf("%d", config.Options.KeepWeekly))
	}

	cmd := exec.Command("restic", args...)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restic cleanup failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// cleanupRclone performs rclone retention cleanup
func cleanupRclone(config *Config) error {
	if config.Retention.KeepDays == 0 {
		return nil
	}

	// Use rclone delete with --min-age to remove old files
	args := []string{
		"delete",
		config.Destination.Remote,
		"--min-age", fmt.Sprintf("%dd", config.Retention.KeepDays),
	}

	cmd := exec.Command("rclone", args...)
	cmd.Env = BaseEnv(config)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rclone cleanup failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// syncWriter serialises writes from concurrent stdout/stderr goroutines.
type syncWriter struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sw *syncWriter) Write(p []byte) (int, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.buf.Write(p)
}

func (sw *syncWriter) String() string {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.buf.String()
}

func runCommandStreaming(name string, args []string, env []string) (string, error) {
	cmd := exec.Command(name, args...)
	if len(env) > 0 {
		cmd.Env = env
	}

	var output syncWriter
	cmd.Stdout = io.MultiWriter(os.Stdout, &output)
	cmd.Stderr = io.MultiWriter(os.Stderr, &output)

	err := cmd.Run()
	return output.String(), err
}
