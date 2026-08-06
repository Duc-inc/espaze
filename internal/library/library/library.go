package library

import (
	"fmt"
	"sync"
	"time"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	"github.com/Duc-inc/espaze/internal/library/game"
	"github.com/Duc-inc/espaze/internal/library/storage"
)

// Library is the in-memory, persisted collection of every game the user
// has scanned in. It's the single source of truth the app bindings and,
// through them, the frontend read from.
type Library struct {
	mu    sync.RWMutex
	store storage.Store
	games []game.Game
}

// New wraps a Store; call Load before using the library.
func New(store storage.Store) *Library {
	return &Library{store: store}
}

// Load reads the persisted library from storage into memory.
func (l *Library) Load() error {
	games, err := l.store.Load()
	if err != nil {
		return fmt.Errorf("library: load: %w", err)
	}
	l.mu.Lock()
	l.games = games
	l.mu.Unlock()
	return nil
}

// ScanFolder walks root for ROMs matching any registered core, adds newly
// found games to the library and persists the result. It returns how many
// new games were added.
func (l *Library) ScanFolder(root string) (added int, err error) {
	extToSystem := core.ExtensionIndex()
	scanned, err := Scan(root, extToSystem)
	if err != nil {
		return 0, fmt.Errorf("library: scan %s: %w", root, err)
	}

	l.mu.Lock()
	before := len(l.games)
	l.games = mergeGames(l.games, scanned)
	added = len(l.games) - before
	snapshot := append([]game.Game(nil), l.games...)
	l.mu.Unlock()

	if err := l.store.Save(snapshot); err != nil {
		return added, fmt.Errorf("library: save: %w", err)
	}
	return added, nil
}

// List returns a copy of every game currently in the library.
func (l *Library) List() []game.Game {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]game.Game, len(l.games))
	copy(out, l.games)
	return out
}

// Get looks up a single game by ID.
func (l *Library) Get(id string) (game.Game, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, g := range l.games {
		if g.ID == id {
			return g, true
		}
	}
	return game.Game{}, false
}

// Remove deletes a game from the library (not from disk) and persists it.
func (l *Library) Remove(id string) error {
	l.mu.Lock()
	idx := indexOf(l.games, id)
	if idx == -1 {
		l.mu.Unlock()
		return fmt.Errorf("library: game %q not found", id)
	}
	l.games = append(l.games[:idx], l.games[idx+1:]...)
	snapshot := append([]game.Game(nil), l.games...)
	l.mu.Unlock()

	return l.store.Save(snapshot)
}

// RecordSession updates play time/last-played for a game and persists it.
func (l *Library) RecordSession(id string, duration time.Duration) error {
	l.mu.Lock()
	idx := indexOf(l.games, id)
	if idx == -1 {
		l.mu.Unlock()
		return fmt.Errorf("library: game %q not found", id)
	}
	l.games[idx].RecordSession(duration)
	snapshot := append([]game.Game(nil), l.games...)
	l.mu.Unlock()

	return l.store.Save(snapshot)
}

func indexOf(games []game.Game, id string) int {
	for i, g := range games {
		if g.ID == id {
			return i
		}
	}
	return -1
}
