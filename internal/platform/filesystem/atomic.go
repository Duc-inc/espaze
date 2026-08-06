package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWriteFile writes data to path without ever leaving a half-written
// file behind: it writes to a temp file in the same directory, then renames
// it into place, which is atomic on every platform this app targets.
func AtomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("filesystem: create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("filesystem: write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("filesystem: close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("filesystem: rename into place: %w", err)
	}
	return nil
}
