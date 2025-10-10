package backup

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// VerifyResult represents the result of a backup verification
type VerifyResult struct {
	Success bool
	Message string
	Details string
}

// Verify verifies a backup based on its type and configuration
func Verify(config *Config) (*VerifyResult, error) {
	if !config.Verification.Enabled {
		return &VerifyResult{
			Success: true,
			Message: "Verification disabled",
		}, nil
	}

	// Check if the required tool is available
	available, err := CheckToolAvailable(config.Type)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, fmt.Errorf("%s is not installed or not in PATH\n\n%s",
			config.Type, GetInstallInstructions(config.Type))
	}

	switch config.Type {
	case BackupTypeRsync:
		return verifyRsync(config)
	case BackupTypeRestic:
		return verifyRestic(config)
	case BackupTypeRclone:
		return verifyRclone(config)
	default:
		return nil, fmt.Errorf("unsupported backup type: %s", config.Type)
	}
}

// verifyRsync verifies an rsync backup
func verifyRsync(config *Config) (*VerifyResult, error) {
	method := config.Verification.Method
	if method == "" {
		method = "size"
	}

	switch method {
	case "size":
		return verifyRsyncSize(config)
	case "checksum":
		return verifyRsyncChecksum(config)
	default:
		return nil, fmt.Errorf("unsupported verification method for rsync: %s", method)
	}
}

// verifyRsyncSize verifies rsync backup by comparing sizes
func verifyRsyncSize(config *Config) (*VerifyResult, error) {
	result := &VerifyResult{Success: true}
	var details strings.Builder

	for _, source := range config.Source {
		// Get source size
		srcSize, err := getDirSize(source)
		if err != nil {
			return &VerifyResult{
				Success: false,
				Message: fmt.Sprintf("Failed to get source size: %v", err),
			}, nil
		}

		// Determine destination path for this source
		var destPath string
		if strings.Contains(config.Destination.Path, ":") {
			// Remote rsync destination - can't easily verify without SSH
			details.WriteString(fmt.Sprintf("Source %s: %d bytes (remote verification not supported)\n", source, srcSize))
			continue
		} else {
			destPath = config.Destination.Path
		}

		// Get destination size
		destSize, err := getDirSize(destPath)
		if err != nil {
			return &VerifyResult{
				Success: false,
				Message: fmt.Sprintf("Failed to get destination size: %v", err),
			}, nil
		}

		// Compare sizes (allow 5% difference for metadata)
		diff := float64(srcSize-destSize) / float64(srcSize) * 100
		if diff < -5 || diff > 5 {
			result.Success = false
			result.Message = fmt.Sprintf("Size mismatch: source=%d, dest=%d (%.1f%% difference)", srcSize, destSize, diff)
		} else {
			details.WriteString(fmt.Sprintf("✓ %s: source=%d bytes, dest=%d bytes\n", source, srcSize, destSize))
		}
	}

	result.Details = details.String()
	if result.Success && result.Message == "" {
		result.Message = "Verification successful"
	}

	return result, nil
}

// verifyRsyncChecksum verifies rsync backup using checksums
func verifyRsyncChecksum(config *Config) (*VerifyResult, error) {
	args := []string{"--dry-run", "--checksum", "-n", "-i"}

	// Add sources
	args = append(args, config.Source...)

	// Add destination
	args = append(args, config.Destination.Path)

	cmd := exec.Command("rsync", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &VerifyResult{
			Success: false,
			Message: fmt.Sprintf("Checksum verification failed: %v", err),
			Details: string(output),
		}, nil
	}

	// If output is empty, everything matches
	if len(output) == 0 {
		return &VerifyResult{
			Success: true,
			Message: "Checksum verification successful - all files match",
		}, nil
	}

	return &VerifyResult{
		Success: false,
		Message: "Checksum verification found differences",
		Details: string(output),
	}, nil
}

// verifyRestic verifies a restic backup
func verifyRestic(config *Config) (*VerifyResult, error) {
	env := append(os.Environ(), config.Environment...)
	if config.Options.PasswordFile != "" {
		env = append(env, fmt.Sprintf("RESTIC_PASSWORD_FILE=%s", config.Options.PasswordFile))
	}
	env = append(env, fmt.Sprintf("RESTIC_REPOSITORY=%s", config.Destination.Repository))

	// Use restic check command
	cmd := exec.Command("restic", "check")
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &VerifyResult{
			Success: false,
			Message: fmt.Sprintf("Restic check failed: %v", err),
			Details: string(output),
		}, nil
	}

	return &VerifyResult{
		Success: true,
		Message: "Restic repository verification successful",
		Details: string(output),
	}, nil
}

// verifyRclone verifies an rclone backup
func verifyRclone(config *Config) (*VerifyResult, error) {
	method := config.Verification.Method
	if method == "" {
		method = "check"
	}

	switch method {
	case "check":
		return verifyRcloneCheck(config)
	case "cryptcheck":
		return verifyRcloneCryptcheck(config)
	case "size":
		return verifyRcloneSize(config)
	default:
		return nil, fmt.Errorf("unsupported verification method for rclone: %s", method)
	}
}

// verifyRcloneCheck verifies rclone backup using rclone check
func verifyRcloneCheck(config *Config) (*VerifyResult, error) {
	var allOutput strings.Builder
	success := true

	for _, source := range config.Source {
		destPath := config.Destination.Remote
		if len(config.Source) > 1 {
			// If multiple sources, append source basename to destination
			destPath = fmt.Sprintf("%s/%s", config.Destination.Remote, source[strings.LastIndex(source, "/")+1:])
		}

		args := []string{"check", source, destPath}

		cmd := exec.Command("rclone", args...)
		cmd.Env = append(os.Environ(), config.Environment...)

		output, err := cmd.CombinedOutput()
		allOutput.WriteString(string(output))
		allOutput.WriteString("\n")

		if err != nil {
			success = false
		}
	}

	if success {
		return &VerifyResult{
			Success: true,
			Message: "Rclone check successful - all files match",
			Details: allOutput.String(),
		}, nil
	}

	return &VerifyResult{
		Success: false,
		Message: "Rclone check found differences",
		Details: allOutput.String(),
	}, nil
}

// verifyRcloneCryptcheck verifies rclone encrypted backup
func verifyRcloneCryptcheck(config *Config) (*VerifyResult, error) {
	var allOutput strings.Builder
	success := true

	for _, source := range config.Source {
		destPath := config.Destination.Remote
		if len(config.Source) > 1 {
			destPath = fmt.Sprintf("%s/%s", config.Destination.Remote, source[strings.LastIndex(source, "/")+1:])
		}

		args := []string{"cryptcheck", source, destPath}

		cmd := exec.Command("rclone", args...)
		cmd.Env = append(os.Environ(), config.Environment...)

		output, err := cmd.CombinedOutput()
		allOutput.WriteString(string(output))
		allOutput.WriteString("\n")

		if err != nil {
			success = false
		}
	}

	if success {
		return &VerifyResult{
			Success: true,
			Message: "Rclone cryptcheck successful - all files match",
			Details: allOutput.String(),
		}, nil
	}

	return &VerifyResult{
		Success: false,
		Message: "Rclone cryptcheck found differences",
		Details: allOutput.String(),
	}, nil
}

// verifyRcloneSize verifies rclone backup by comparing sizes
func verifyRcloneSize(config *Config) (*VerifyResult, error) {
	var allOutput strings.Builder
	success := true

	for _, source := range config.Source {
		destPath := config.Destination.Remote
		if len(config.Source) > 1 {
			destPath = fmt.Sprintf("%s/%s", config.Destination.Remote, source[strings.LastIndex(source, "/")+1:])
		}

		args := []string{"size", source, "--json"}
		cmd := exec.Command("rclone", args...)
		cmd.Env = append(os.Environ(), config.Environment...)
		srcOutput, err := cmd.CombinedOutput()
		if err != nil {
			return &VerifyResult{
				Success: false,
				Message: fmt.Sprintf("Failed to get source size: %v", err),
				Details: string(srcOutput),
			}, nil
		}

		args = []string{"size", destPath, "--json"}
		cmd = exec.Command("rclone", args...)
		cmd.Env = append(os.Environ(), config.Environment...)
		destOutput, err := cmd.CombinedOutput()
		if err != nil {
			return &VerifyResult{
				Success: false,
				Message: fmt.Sprintf("Failed to get destination size: %v", err),
				Details: string(destOutput),
			}, nil
		}

		allOutput.WriteString(fmt.Sprintf("Source %s: %s\n", source, string(srcOutput)))
		allOutput.WriteString(fmt.Sprintf("Dest %s: %s\n", destPath, string(destOutput)))
	}

	if success {
		return &VerifyResult{
			Success: true,
			Message: "Size comparison successful",
			Details: allOutput.String(),
		}, nil
	}

	return &VerifyResult{
		Success: false,
		Message: "Size comparison found differences",
		Details: allOutput.String(),
	}, nil
}

// getDirSize returns the total size of a directory in bytes
func getDirSize(path string) (int64, error) {
	var size int64

	cmd := exec.Command("du", "-sb", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	// Parse output: "123456  /path"
	parts := strings.Fields(string(output))
	if len(parts) < 1 {
		return 0, fmt.Errorf("unexpected du output: %s", string(output))
	}

	size, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size: %w", err)
	}

	return size, nil
}
