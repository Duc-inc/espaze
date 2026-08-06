package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Duc-inc/espaze/internal/library/game"
	"github.com/Duc-inc/espaze/internal/platform/filesystem"
)

// JSONStore persists the library as a single JSON array on disk.
type JSONStore struct {
	path string
}

// NewJSONStore returns a Store backed by the file at path.
func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

// Load implements Store. A missing file is treated as an empty library.
func (s *JSONStore) Load() ([]game.Game, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return []game.Game{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("jsonstore: read %s: %w", s.path, err)
	}

	var games []game.Game
	if err := json.Unmarshal(data, &games); err != nil {
		return nil, fmt.Errorf("jsonstore: parse %s: %w", s.path, err)
	}
	return games, nil
}

// Save implements Store, writing atomically so a crash mid-write can't
// corrupt the library file.
func (s *JSONStore) Save(games []game.Game) error {
	data, err := json.MarshalIndent(games, "", "  ")
	if err != nil {
		return fmt.Errorf("jsonstore: encode: %w", err)
	}
	return filesystem.AtomicWriteFile(s.path, data)
}
