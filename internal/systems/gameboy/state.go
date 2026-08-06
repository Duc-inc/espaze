package gameboy

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/gameboy/cpu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/joypad"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/memory"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/ppu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/timer"
)

// snapshot combines every component's own snapshot into one save state.
// The cartridge ROM itself is never included - LoadState assumes the
// same ROM is already loaded, exactly like every other core here.
type snapshot struct {
	CPU    cpu.Snapshot
	MMU    memory.Snapshot
	MBC    memory.MBCSnapshot
	PPU    ppu.Snapshot
	Timer  timer.Snapshot
	Joypad joypad.Snapshot
}

// SaveState implements core.Core.
func (gb *GameBoy) SaveState() ([]byte, error) {
	if !gb.loaded {
		return nil, fmt.Errorf("gameboy: no rom loaded")
	}

	snap := snapshot{
		CPU:    gb.proc.Snapshot(),
		MMU:    gb.mmu.Snapshot(),
		MBC:    gb.mbc.Snapshot(),
		PPU:    gb.video.Snapshot(),
		Timer:  gb.tmr.Snapshot(),
		Joypad: gb.pad.Snapshot(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("gameboy: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (gb *GameBoy) LoadState(data []byte) error {
	if !gb.loaded {
		return fmt.Errorf("gameboy: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("gameboy: load state: %w", err)
	}

	gb.proc.Restore(snap.CPU)
	gb.mmu.Restore(snap.MMU)
	gb.mbc.Restore(snap.MBC)
	gb.video.Restore(snap.PPU)
	gb.tmr.Restore(snap.Timer)
	gb.pad.Restore(snap.Joypad)
	return nil
}
