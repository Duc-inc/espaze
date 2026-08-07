package snes

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/snes/audio"
	"github.com/Duc-inc/espaze/internal/systems/snes/cpu"
	"github.com/Duc-inc/espaze/internal/systems/snes/dsp"
	"github.com/Duc-inc/espaze/internal/systems/snes/memory"
	"github.com/Duc-inc/espaze/internal/systems/snes/ppu"
	"github.com/Duc-inc/espaze/internal/systems/snes/spc700"
)

// snapshot combines every component's own snapshot into one save
// state. The cartridge ROM itself is never included - LoadState
// assumes the same ROM is already loaded, exactly like every other
// core here.
type snapshot struct {
	CPU     cpu.Snapshot
	Bus     memory.Snapshot
	PPU     ppu.Snapshot
	DSP     dsp.Snapshot
	SPC     spc700.Snapshot
	SPCBus  audio.Snapshot
	Ports   audio.PortsSnapshot
	SPCLeft float64
}

// SaveState implements core.Core.
func (s *SNES) SaveState() ([]byte, error) {
	if !s.loaded {
		return nil, fmt.Errorf("snes: no rom loaded")
	}

	snap := snapshot{
		CPU: s.proc.Snapshot(), Bus: s.bus.Snapshot(), PPU: s.video.Snapshot(),
		DSP: s.sound.Snapshot(), SPC: s.spc.Snapshot(), SPCBus: s.spcBus.Snapshot(),
		Ports: s.ports.Snapshot(), SPCLeft: s.spcLeft,
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("snes: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (s *SNES) LoadState(data []byte) error {
	if !s.loaded {
		return fmt.Errorf("snes: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("snes: load state: %w", err)
	}

	s.proc.Restore(snap.CPU)
	s.bus.Restore(snap.Bus)
	s.video.Restore(snap.PPU)
	s.sound.Restore(snap.DSP)
	s.spc.Restore(snap.SPC)
	s.spcBus.Restore(snap.SPCBus)
	s.ports.Restore(snap.Ports)
	s.spcLeft = snap.SPCLeft
	return nil
}
