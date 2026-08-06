package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Duc-inc/espaze/internal/platform/filesystem"
)

// Config holds every user-adjustable setting for the app.
type Config struct {
	LibraryFolders []string `json:"libraryFolders"`
	WindowWidth    int      `json:"windowWidth,omitempty"`
	WindowHeight   int      `json:"windowHeight,omitempty"`
}

// Load reads settings from disk, returning defaults if the file is missing.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	return &cfg, nil
}

// Save writes settings to disk, creating parent directories as needed.
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("config: encode: %w", err)
	}
	return filesystem.AtomicWriteFile(path, data)
}

// AddLibraryFolder registers a folder to scan for ROMs, ignoring duplicates.
func (c *Config) AddLibraryFolder(path string) {
	for _, existing := range c.LibraryFolders {
		if existing == path {
			return
		}
	}
	c.LibraryFolders = append(c.LibraryFolders, path)
}

// RemoveLibraryFolder stops a folder from being scanned in the future.
// Games already found there stay in the library until the caller also
// removes them (see library.Library.RemoveByPathPrefix).
func (c *Config) RemoveLibraryFolder(path string) {
	for i, existing := range c.LibraryFolders {
		if existing == path {
			c.LibraryFolders = append(c.LibraryFolders[:i], c.LibraryFolders[i+1:]...)
			return
		}
	}
}
