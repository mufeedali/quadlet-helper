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

// Restart multiple systemd user units.
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
func IsActive(unit string) bool {
	allArgs := []string{"--user", "is-active", unit}
	cmd := exec.Command("systemctl", allArgs...)
	return cmd.Run() == nil
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
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
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
