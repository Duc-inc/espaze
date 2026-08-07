package ngpc

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/ngpc/audio"
	"github.com/Duc-inc/espaze/internal/systems/ngpc/cpu"
	"github.com/Duc-inc/espaze/internal/systems/ngpc/memory"
	"github.com/Duc-inc/espaze/internal/systems/ngpc/ppu"
	sms "github.com/Duc-inc/espaze/internal/systems/sms/cpu"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
)

// snapshot combines every component's own snapshot into one save
// state. The cartridge ROM itself is never included - LoadState
// assumes the same ROM is already loaded, exactly like every other
// core here.
type snapshot struct {
	CPU     cpu.Snapshot
	Bus     memory.Snapshot
	PPU     ppu.Snapshot
	PSG     psg.Snapshot
	Z80     sms.Snapshot
	Z80Bus  audio.Snapshot
	Z80Left float64
}

// SaveState implements core.Core.
func (n *NGPC) SaveState() ([]byte, error) {
	if !n.loaded {
		return nil, fmt.Errorf("ngpc: no rom loaded")
	}

	snap := snapshot{
		CPU: n.proc.Snapshot(), Bus: n.bus.Snapshot(), PPU: n.video.Snapshot(),
		PSG: n.sound.Snapshot(), Z80: n.z80.Snapshot(), Z80Bus: n.z80bus.Snapshot(),
		Z80Left: n.z80Left,
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("ngpc: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (n *NGPC) LoadState(data []byte) error {
	if !n.loaded {
		return fmt.Errorf("ngpc: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("ngpc: load state: %w", err)
	}

	n.proc.Restore(snap.CPU)
	n.bus.Restore(snap.Bus)
	n.video.Restore(snap.PPU)
	n.sound.Restore(snap.PSG)
	n.z80.Restore(snap.Z80)
	n.z80bus.Restore(snap.Z80Bus)
	n.z80Left = snap.Z80Left
	return nil
}
