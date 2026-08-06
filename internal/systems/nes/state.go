package nes

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/nes/apu"
	"github.com/Duc-inc/espaze/internal/systems/nes/cpu"
	"github.com/Duc-inc/espaze/internal/systems/nes/memory"
	"github.com/Duc-inc/espaze/internal/systems/nes/ppu"
)

// snapshot combines every component's own snapshot into one save state.
// The cartridge ROM itself is never included - LoadState assumes the
// same ROM is already loaded, exactly like every other core here.
type snapshot struct {
	CPU cpu.Snapshot
	PPU ppu.Snapshot
	APU apu.Snapshot
	Bus memory.Snapshot
}

// SaveState implements core.Core.
func (n *NES) SaveState() ([]byte, error) {
	if !n.loaded {
		return nil, fmt.Errorf("nes: no rom loaded")
	}

	snap := snapshot{
		CPU: n.proc.Snapshot(),
		PPU: n.video.Snapshot(),
		APU: n.sound.Snapshot(),
		Bus: n.bus.Snapshot(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("nes: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (n *NES) LoadState(data []byte) error {
	if !n.loaded {
		return fmt.Errorf("nes: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("nes: load state: %w", err)
	}

	n.proc.Restore(snap.CPU)
	n.video.Restore(snap.PPU)
	n.sound.Restore(snap.APU)
	return n.bus.Restore(snap.Bus)
}
