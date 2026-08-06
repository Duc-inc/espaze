package gbc

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/gameboy/apu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/cpu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/joypad"
	dmgmem "github.com/Duc-inc/espaze/internal/systems/gameboy/memory"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/timer"
	"github.com/Duc-inc/espaze/internal/systems/gbc/memory"
	"github.com/Duc-inc/espaze/internal/systems/gbc/ppu"
)

// snapshot combines every component's own snapshot into one save state.
// The cartridge ROM itself is never included - LoadState assumes the
// same ROM is already loaded, exactly like every other core here.
type snapshot struct {
	CPU    cpu.Snapshot
	MMU    memory.Snapshot
	MBC    dmgmem.MBCSnapshot
	PPU    ppu.Snapshot
	Timer  timer.Snapshot
	Joypad joypad.Snapshot
	APU    apu.Snapshot
}

// SaveState implements core.Core.
func (gb *GBC) SaveState() ([]byte, error) {
	if !gb.loaded {
		return nil, fmt.Errorf("gbc: no rom loaded")
	}

	snap := snapshot{
		CPU:    gb.proc.Snapshot(),
		MMU:    gb.mmu.Snapshot(),
		MBC:    gb.mbc.Snapshot(),
		PPU:    gb.video.Snapshot(),
		Timer:  gb.tmr.Snapshot(),
		Joypad: gb.pad.Snapshot(),
		APU:    gb.sound.Snapshot(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("gbc: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (gb *GBC) LoadState(data []byte) error {
	if !gb.loaded {
		return fmt.Errorf("gbc: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("gbc: load state: %w", err)
	}

	gb.proc.Restore(snap.CPU)
	gb.mmu.Restore(snap.MMU)
	gb.mbc.Restore(snap.MBC)
	gb.video.Restore(snap.PPU)
	gb.tmr.Restore(snap.Timer)
	gb.pad.Restore(snap.Joypad)
	gb.sound.Restore(snap.APU)
	return nil
}
