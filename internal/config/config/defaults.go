package config

// Default returns a Config with no library folders configured yet -
// the user adds them from the frontend on first run.
func Default() *Config {
	return &Config{
		LibraryFolders: []string{},
	}
}
