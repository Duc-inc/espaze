package filesystem

import (
	"fmt"
	"os"
)

// EnsureDir creates a directory (and any missing parents) if it doesn't
// already exist, matching the permissions used across the app's data dirs.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("filesystem: create dir %s: %w", path, err)
	}
	return nil
}

// Exists reports whether a path is present on disk, regardless of type.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir reports whether path exists and is a directory.
func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
