package gba

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/gba/apu"
	"github.com/Duc-inc/espaze/internal/systems/gba/cpu"
	"github.com/Duc-inc/espaze/internal/systems/gba/memory"
	"github.com/Duc-inc/espaze/internal/systems/gba/ppu"
)

// snapshot combines every component's own snapshot into one save
// state. The cartridge ROM itself is never included - LoadState
// assumes the same ROM is already loaded, exactly like every other
// core here.
type snapshot struct {
	CPU cpu.Snapshot
	Bus memory.Snapshot
	PPU ppu.Snapshot
	APU apu.Snapshot
}

// SaveState implements core.Core.
func (g *GBA) SaveState() ([]byte, error) {
	if !g.loaded {
		return nil, fmt.Errorf("gba: no rom loaded")
	}

	snap := snapshot{
		CPU: g.proc.Snapshot(),
		Bus: g.bus.Snapshot(),
		PPU: g.video.Snapshot(),
		APU: g.sound.Snapshot(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("gba: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (g *GBA) LoadState(data []byte) error {
	if !g.loaded {
		return fmt.Errorf("gba: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("gba: load state: %w", err)
	}

	g.proc.Restore(snap.CPU)
	g.bus.Restore(snap.Bus)
	g.video.Restore(snap.PPU)
	g.sound.Restore(snap.APU)
	return nil
}
