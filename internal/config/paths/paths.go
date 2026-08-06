package paths

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Duc-inc/espaze/internal/platform/filesystem"
)

// appDirName is the folder created under the OS config/data directory.
const appDirName = "espaze"

// DataDir returns the per-user directory this app stores everything in:
// %AppData%/espaze on Windows, ~/Library/Application Support/espaze on
// macOS, ~/.config/espaze on Linux (via os.UserConfigDir).
func DataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("paths: resolve user config dir: %w", err)
	}
	return filepath.Join(base, appDirName), nil
}

// ConfigFile is where the app's settings (library folders, etc.) live.
func ConfigFile() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LibraryStoreFile is where the scanned game library is persisted.
func LibraryStoreFile() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "library.json"), nil
}

// ArtworkDir is where downloaded/cached cover art is cached on disk.
func ArtworkDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "artwork"), nil
}

// SaveStateDir is where per-game save states are written.
func SaveStateDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "savestates"), nil
}

// EnsureDataDirs creates every directory the app expects to exist.
func EnsureDataDirs() error {
	dir, err := DataDir()
	if err != nil {
		return err
	}
	art, err := ArtworkDir()
	if err != nil {
		return err
	}
	states, err := SaveStateDir()
	if err != nil {
		return err
	}
	for _, d := range []string{dir, art, states} {
		if err := filesystem.EnsureDir(d); err != nil {
			return err
		}
	}
	return nil
}
