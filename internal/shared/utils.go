package shared

import (
	"fmt"
	"io/fs"
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
func WalkWithSymlinks(root string, walkFn func(path string, d fs.DirEntry) error) error {
	visited := make(map[string]bool)

	var walk func(string) error
	walk = func(currentRoot string) error {
		return filepath.WalkDir(currentRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// We only resolve paths for directories to detect loops.
			// Regular files and symlinks to files are processed as-is (allowing duplicates/aliases).
			if d.IsDir() {
				realPath, err := filepath.EvalSymlinks(path)
				if err != nil {
					realPath = path
				}

				if visited[realPath] {
					return filepath.SkipDir
				}
				visited[realPath] = true
			}

			// If it's a symlink, checks if it points to a directory we should traverse
			if d.Type()&os.ModeSymlink != 0 {
				target, err := os.Readlink(path)
				if err == nil {
					// Make absolute if relative
					if !filepath.IsAbs(target) {
						target = filepath.Join(filepath.Dir(path), target)
					}
					targetInfo, err := os.Stat(target)
					if err == nil && targetInfo.IsDir() {
						// Recurse into the symlinked directory
						// We ignore the error from walk() to continue processing the current directory
						_ = walk(target)
					}
				}
			}

			return walkFn(path, d)
		})
	}

	return walk(root)
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
