package backup

import (
	"encoding/json"
	"fmt"
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
	normalized := config.Normalized()
	config = &normalized

	if !config.Verification.Enabled {
		return &VerifyResult{
			Success: true,
			Message: "Verification disabled",
		}, nil
	}

	// Validate config for type-specific rules before running tools
	if err := config.Validate(); err != nil {
		return nil, err
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
	method := string(config.Verification.Method)
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
			fmt.Fprintf(&details, "Source %s: %d bytes (remote verification not supported)\n", source, srcSize)
			continue
		} else {
			destPath = RsyncDestPath(config.Destination.Path, source, len(config.Source))
		}

		// Get destination size
		destSize, err := getDirSize(destPath)
		if err != nil {
			return &VerifyResult{
				Success: false,
				Message: fmt.Sprintf("Failed to get destination size: %v", err),
			}, nil
		}

		if srcSize == 0 {
			if destSize != 0 {
				result.Success = false
				result.Message = fmt.Sprintf("Size mismatch: source=%d, dest=%d", srcSize, destSize)
			} else {
				fmt.Fprintf(&details, "✓ %s: source=0 bytes, dest=0 bytes\n", source)
			}
			continue
		}

		// Compare sizes (allow 5% difference for metadata)
		diff := float64(srcSize-destSize) / float64(srcSize) * 100
		if diff < -5 || diff > 5 {
			result.Success = false
			result.Message = fmt.Sprintf("Size mismatch: source=%d, dest=%d (%.1f%% difference)", srcSize, destSize, diff)
		} else {
			fmt.Fprintf(&details, "✓ %s: source=%d bytes, dest=%d bytes\n", source, srcSize, destSize)
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
	env := ResticEnv(config)

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
	switch config.Verification.Method {
	case VerificationMethodCheck, VerificationMethodCryptCheck:
		return runRcloneCheck(config, config.Verification.Method)
	case VerificationMethodSize:
		return verifyRcloneSize(config)
	default:
		// Shouldn't happen - validated earlier
		return nil, fmt.Errorf("unsupported verification method for rclone: %s", config.Verification.Method)
	}
}

// runRcloneCheck runs rclone check/cryptcheck for each source.
func runRcloneCheck(config *Config, verb VerificationMethod) (*VerifyResult, error) {
	v := string(verb)
	var allOutput strings.Builder
	success := true

	for _, source := range config.Source {
		destPath := RcloneDestPath(config.Destination.Remote, source, len(config.Source))

		args := []string{v, source, destPath}
		cmd := exec.Command("rclone", args...)
		cmd.Env = BaseEnv(config)

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
			Message: fmt.Sprintf("Rclone %s successful - all files match", v),
			Details: allOutput.String(),
		}, nil
	}

	return &VerifyResult{
		Success: false,
		Message: fmt.Sprintf("Rclone %s found differences", v),
		Details: allOutput.String(),
	}, nil
}

// verifyRcloneSize verifies rclone backup by comparing sizes
func verifyRcloneSize(config *Config) (*VerifyResult, error) {
	var allOutput strings.Builder
	success := true

	type rcloneSize struct {
		Count int64 `json:"count"`
		Bytes int64 `json:"bytes"`
	}

	for _, source := range config.Source {
		destPath := RcloneDestPath(config.Destination.Remote, source, len(config.Source))

		// source
		args := []string{"size", source, "--json"}
		cmd := exec.Command("rclone", args...)
		cmd.Env = BaseEnv(config)
		srcOutput, err := cmd.CombinedOutput()
		if err != nil {
			return &VerifyResult{Success: false, Message: fmt.Sprintf("Failed to get source size: %v", err), Details: string(srcOutput)}, nil
		}

		var src rcloneSize
		if err := json.Unmarshal(srcOutput, &src); err != nil {
			return &VerifyResult{Success: false, Message: fmt.Sprintf("Failed to parse source size JSON: %v", err), Details: string(srcOutput)}, nil
		}

		// dest
		args = []string{"size", destPath, "--json"}
		cmd = exec.Command("rclone", args...)
		cmd.Env = BaseEnv(config)
		destOutput, err := cmd.CombinedOutput()
		if err != nil {
			return &VerifyResult{Success: false, Message: fmt.Sprintf("Failed to get destination size: %v", err), Details: string(destOutput)}, nil
		}

		var dst rcloneSize
		if err := json.Unmarshal(destOutput, &dst); err != nil {
			return &VerifyResult{Success: false, Message: fmt.Sprintf("Failed to parse destination size JSON: %v", err), Details: string(destOutput)}, nil
		}

		// compare with 5% tolerance
		match := true
		var detailLine string
		if src.Bytes == 0 {
			if dst.Bytes != 0 {
				match = false
				detailLine = fmt.Sprintf("%s: source=0 bytes, dest=%d bytes (mismatch)", source, dst.Bytes)
			} else {
				detailLine = fmt.Sprintf("%s: source=0 bytes, dest=0 bytes", source)
			}
		} else {
			diff := float64(src.Bytes-dst.Bytes) / float64(src.Bytes) * 100
			if diff < -5 || diff > 5 {
				match = false
				detailLine = fmt.Sprintf("%s: source=%d bytes, dest=%d bytes (%.1f%% diff)", source, src.Bytes, dst.Bytes, diff)
			} else {
				detailLine = fmt.Sprintf("✓ %s: source=%d bytes, dest=%d bytes (%.1f%% diff)", source, src.Bytes, dst.Bytes, diff)
			}
		}

		if !match {
			success = false
		}

		allOutput.WriteString(detailLine)
		allOutput.WriteString("\n")
		// include raw JSON for debugging
		fmt.Fprintf(&allOutput, "Source JSON: %s\n", string(srcOutput))
		fmt.Fprintf(&allOutput, "Dest JSON: %s\n", string(destOutput))
	}

	if success {
		return &VerifyResult{Success: true, Message: "Size comparison successful", Details: allOutput.String()}, nil
	}

	return &VerifyResult{Success: false, Message: "Size comparison found differences", Details: allOutput.String()}, nil
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
