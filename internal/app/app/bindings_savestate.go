package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Duc-inc/espaze/internal/config/paths"
	"github.com/Duc-inc/espaze/internal/platform/filesystem"
)

// maxSaveSlots is how many save slots each game gets. A small fixed
// number keeps the UI simple (no slot management, just three buttons).
const maxSaveSlots = 3

// SaveSlotInfo describes one save slot for the frontend: whether it has
// been used yet, and if so, when it was last written.
type SaveSlotInfo struct {
	Slot    int        `json:"slot"`
	SavedAt *time.Time `json:"savedAt,omitempty"`
}

// SaveStateToSlot captures the running game's state into a numbered slot
// (0..maxSaveSlots-1), persisted to disk so it survives app restarts.
func (a *App) SaveStateToSlot(slot int) error {
	if a.engine == nil {
		return fmt.Errorf("app: no game running")
	}
	if err := validateSlot(slot); err != nil {
		return err
	}

	data, err := a.engine.SaveState()
	if err != nil {
		return fmt.Errorf("app: save state: %w", err)
	}

	path, err := a.slotPath(a.activeGameID, slot)
	if err != nil {
		return err
	}
	return filesystem.AtomicWriteFile(path, data)
}

// LoadStateFromSlot restores a previously saved slot into the running game.
func (a *App) LoadStateFromSlot(slot int) error {
	if a.engine == nil {
		return fmt.Errorf("app: no game running")
	}
	if err := validateSlot(slot); err != nil {
		return err
	}

	path, err := a.slotPath(a.activeGameID, slot)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("app: read save slot %d: %w", slot, err)
	}
	return a.engine.LoadState(data)
}

// ListSaveSlots reports every slot's state for the currently running game.
func (a *App) ListSaveSlots() ([]SaveSlotInfo, error) {
	if a.activeGameID == "" {
		return nil, fmt.Errorf("app: no game running")
	}

	slots := make([]SaveSlotInfo, maxSaveSlots)
	for i := range slots {
		slots[i] = SaveSlotInfo{Slot: i}
	}

	dir, err := a.slotDir(a.activeGameID)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return slots, nil
	}
	if err != nil {
		return nil, fmt.Errorf("app: list save slots: %w", err)
	}

	for _, entry := range entries {
		var slot int
		if _, err := fmt.Sscanf(entry.Name(), "slot-%d.state", &slot); err != nil {
			continue
		}
		if slot < 0 || slot >= maxSaveSlots {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		savedAt := info.ModTime()
		slots[slot].SavedAt = &savedAt
	}

	sort.Slice(slots, func(i, j int) bool { return slots[i].Slot < slots[j].Slot })
	return slots, nil
}

func validateSlot(slot int) error {
	if slot < 0 || slot >= maxSaveSlots {
		return fmt.Errorf("app: slot %d out of range (0-%d)", slot, maxSaveSlots-1)
	}
	return nil
}

func (a *App) slotDir(gameID string) (string, error) {
	base, err := paths.SaveStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, gameID), nil
}

func (a *App) slotPath(gameID string, slot int) (string, error) {
	dir, err := a.slotDir(gameID)
	if err != nil {
		return "", err
	}
	if err := filesystem.EnsureDir(dir); err != nil {
		return "", err
	}
	return filepath.Join(dir, fmt.Sprintf("slot-%d.state", slot)), nil
}
