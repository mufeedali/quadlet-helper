package systemd

import (
	"fmt"
	"os/exec"
	"strings"
)

// runSystemctl executes a systemctl command with the --user flag.
func runSystemctl(args ...string) (string, error) {
	allArgs := append([]string{"--user"}, args...)
	cmd := exec.Command("systemctl", allArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("error running systemctl %s: %w", strings.Join(args, " "), err)
	}
	return string(output), nil
}

// DaemonReload reloads the systemd user daemon.
func DaemonReload() (string, error) {
	return runSystemctl("daemon-reload")
}

// Start starts a systemd user unit.
func Start(unit string) (string, error) {
	return runSystemctl("start", unit)
}

// Stop stops a systemd user unit.
func Stop(unit string) (string, error) {
	return runSystemctl("stop", unit)
}

// StopMultiple stops multiple systemd user units.
func StopMultiple(units []string) (string, error) {
	args := append([]string{"stop"}, units...)
	return runSystemctl(args...)
}

// StartMultiple starts multiple systemd user units.
func StartMultiple(units []string) (string, error) {
	args := append([]string{"start"}, units...)
	return runSystemctl(args...)
}

// Enable enables a systemd user unit to start on boot.
func Enable(unit string) (string, error) {
	return runSystemctl("enable", unit)
}

// Disable disables a systemd user unit from starting on boot.
func Disable(unit string) (string, error) {
	return runSystemctl("disable", unit)
}

// Restart restarts a systemd user unit.
func Restart(unit string) (string, error) {
	return runSystemctl("restart", unit)
}

// RestartMultiple restarts multiple systemd user units.
func RestartMultiple(units []string) (string, error) {
	args := append([]string{"restart"}, units...)
	return runSystemctl(args...)
}

// Status gets the status of one or more systemd user units.
// This function does not return an error on non-zero exit codes from systemctl,
// as systemctl status returns a non-zero code for inactive units.
func Status(units ...string) (string, error) {
	allArgs := append([]string{"--user", "--no-pager", "status"}, units...)
	cmd := exec.Command("systemctl", allArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ListTimers lists systemd timers.
func ListTimers(timer string) (string, error) {
	allArgs := []string{"--user", "list-timers", timer, "--no-pager"}
	cmd := exec.Command("systemctl", allArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// IsActive checks if a systemd user unit is active.
func IsActive(unit string) (bool, error) {
	allArgs := []string{"--user", "is-active", unit}
	cmd := exec.Command("systemctl", allArgs...)
	output, err := cmd.CombinedOutput()
	return parseIsActiveResult(unit, string(output), err)
}

func parseIsActiveResult(unit, output string, err error) (bool, error) {
	status := strings.TrimSpace(output)
	switch status {
	case "active":
		return true, nil
	case "inactive", "failed", "activating", "deactivating", "reloading", "refreshing", "maintenance", "unknown", "not-found":
		return false, nil
	case "":
		if err != nil {
			return false, fmt.Errorf("error checking active state for %s: %w", unit, err)
		}
		return false, fmt.Errorf("unexpected empty active state for %s", unit)
	default:
		if err != nil {
			return false, fmt.Errorf("error checking active state for %s: %w\n%s", unit, err, status)
		}
		return false, fmt.Errorf("unexpected active state for %s: %s", unit, status)
	}
}

// IsActiveMultiple checks if multiple systemd user units are active.
// Returns a slice of active statuses corresponding to the input units.
func IsActiveMultiple(units []string) ([]bool, error) {
	if len(units) == 0 {
		return nil, nil
	}

	allArgs := append([]string{"--user", "is-active"}, units...)
	cmd := exec.Command("systemctl", allArgs...)
	output, err := cmd.CombinedOutput()
	return parseIsActiveMultipleResult(units, string(output), err)
}

func parseIsActiveMultipleResult(units []string, output string, err error) ([]bool, error) {
	trimmedOutput := strings.TrimSpace(output)
	if trimmedOutput == "" {
		if err != nil {
			return nil, fmt.Errorf("error checking active state for units: %w", err)
		}
		return nil, fmt.Errorf("unexpected empty active state output")
	}

	lines := strings.Split(trimmedOutput, "\n")
	if len(lines) != len(units) {
		if err != nil {
			return nil, fmt.Errorf("error checking active state for units: %w\n%s", err, trimmedOutput)
		}
		return nil, fmt.Errorf("unexpected number of active states: got %d, want %d\n%s", len(lines), len(units), trimmedOutput)
	}

	result := make([]bool, len(units))

	for i, unit := range units {
		active, parseErr := parseIsActiveResult(unit, lines[i], nil)
		if parseErr != nil {
			if err != nil {
				return nil, fmt.Errorf("error checking active state for %s: %w\n%s", unit, err, strings.TrimSpace(lines[i]))
			}
			return nil, parseErr
		}
		result[i] = active
	}

	return result, nil
}

// ListActiveServices returns a list of all active systemd user services.
func ListActiveServices() ([]string, error) {
	allArgs := []string{"--user", "list-units", "--type=service", "--state=active", "--no-legend", "--plain"}
	cmd := exec.Command("systemctl", allArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing active services: %w", err)
	}

	var services []string
	lines := strings.SplitSeq(string(output), "\n")
	for line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			services = append(services, fields[0])
		}
	}
	return services, nil
}

// Show gets properties of a systemd user unit.
func Show(unit, property string) (string, error) {
	allArgs := []string{"--user", "show", unit, "--property=" + property, "--value"}
	cmd := exec.Command("systemctl", allArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
