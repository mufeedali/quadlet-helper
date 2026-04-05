package systemd

import (
	"fmt"
	"os"
	"path/filepath"
)

type UserUnitFile struct {
	Name    string
	Content string
	Mode    os.FileMode
}

type UninstallResult struct {
	RemovedPaths []string
	Warnings     []error
}

func UserDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}
	return filepath.Join(home, ".config", "systemd", "user"), nil
}

func InstallUserUnits(files []UserUnitFile, activateUnits []string) ([]string, error) {
	userDir, err := UserDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("creating systemd directory: %w", err)
	}

	paths := make([]string, 0, len(files))
	removeWritten := func() {
		for _, path := range paths {
			_ = os.Remove(path)
		}
	}

	for _, file := range files {
		mode := file.Mode
		if mode == 0 {
			mode = 0644
		}
		path := filepath.Join(userDir, file.Name)
		if err := os.WriteFile(path, []byte(file.Content), mode); err != nil {
			removeWritten()
			return nil, fmt.Errorf("writing %s: %w", file.Name, err)
		}
		paths = append(paths, path)
	}

	if _, err := DaemonReload(); err != nil {
		removeWritten()
		return nil, fmt.Errorf("reloading systemd: %w", err)
	}
	for _, unit := range activateUnits {
		if _, err := Enable(unit); err != nil {
			removeWritten()
			_, _ = DaemonReload()
			return nil, fmt.Errorf("enabling %s: %w", unit, err)
		}
		if _, err := Start(unit); err != nil {
			removeWritten()
			_, _ = DaemonReload()
			return nil, fmt.Errorf("starting %s: %w", unit, err)
		}
	}

	return paths, nil
}

func UninstallUserUnits(stopUnits []string, disableUnits []string, removeFiles []string) (*UninstallResult, error) {
	userDir, err := UserDir()
	if err != nil {
		return nil, err
	}

	result := &UninstallResult{}
	for _, unit := range stopUnits {
		if _, err := Stop(unit); err != nil {
			result.Warnings = append(result.Warnings, fmt.Errorf("stopping %s: %w", unit, err))
		}
	}
	for _, unit := range disableUnits {
		if _, err := Disable(unit); err != nil {
			result.Warnings = append(result.Warnings, fmt.Errorf("disabling %s: %w", unit, err))
		}
	}
	for _, fileName := range removeFiles {
		path := filepath.Join(userDir, fileName)
		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				result.Warnings = append(result.Warnings, fmt.Errorf("removing %s: %w", fileName, err))
			}
			continue
		}
		result.RemovedPaths = append(result.RemovedPaths, path)
	}
	if _, err := DaemonReload(); err != nil {
		result.Warnings = append(result.Warnings, fmt.Errorf("reloading systemd: %w", err))
	}

	return result, nil
}
