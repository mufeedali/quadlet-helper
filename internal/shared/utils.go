package shared

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveContainersDir resolves symlinks in the containers directory path.
// Returns the resolved path, or the original path if resolution fails (with a warning).
func ResolveContainersDir(containersDir string) string {
	realPath, err := filepath.EvalSymlinks(containersDir)
	if err != nil {
		fmt.Println(WarningStyle.Render(fmt.Sprintf("Warning: could not resolve symlink for %s: %v. Proceeding with original path.", containersDir, err)))
		return containersDir
	}
	return realPath
}

// WalkWithSymlinks walks a directory tree following symlinks.
// It prevents infinite loops by tracking visited directories.
// The walkFn is called for each file and directory found.
func WalkWithSymlinks(root string, walkFn func(path string, info os.FileInfo) error) error {
	visited := make(map[string]bool)

	var walk func(string) error
	walk = func(currentRoot string) error {
		return filepath.Walk(currentRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get real path to detect circular symlinks
			realPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				realPath = path
			}

			// Skip if already visited (prevents infinite loops)
			if visited[realPath] {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			visited[realPath] = true

			// If it's a symlink to a directory, follow it
			if info.Mode()&os.ModeSymlink != 0 {
				target, err := os.Readlink(path)
				if err == nil {
					// Make absolute if relative
					if !filepath.IsAbs(target) {
						target = filepath.Join(filepath.Dir(path), target)
					}
					targetInfo, err := os.Stat(target)
					if err == nil && targetInfo.IsDir() {
						return walk(target)
					}
				}
			}

			return walkFn(path, info)
		})
	}

	return walk(root)
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
