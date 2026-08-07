package genesis

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/genesis/audio"
	"github.com/Duc-inc/espaze/internal/systems/genesis/cpu"
	"github.com/Duc-inc/espaze/internal/systems/genesis/memory"
	"github.com/Duc-inc/espaze/internal/systems/genesis/vdp"
	"github.com/Duc-inc/espaze/internal/systems/genesis/ym2612"
	sms "github.com/Duc-inc/espaze/internal/systems/sms/cpu"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
)

// snapshot combines every component's own snapshot into one save
// state. The cartridge ROM itself is never included - LoadState
// assumes the same ROM is already loaded, exactly like every other
// core here.
type snapshot struct {
	CPU          cpu.Snapshot
	Bus          memory.Snapshot
	VDP          vdp.Snapshot
	YM2612       ym2612.Snapshot
	PSG          psg.Snapshot
	Z80          sms.Snapshot
	Z80Bus       audio.Snapshot
	Z80Left      float64
	PrevZ80Reset bool
}

// SaveState implements core.Core.
func (g *Genesis) SaveState() ([]byte, error) {
	if !g.loaded {
		return nil, fmt.Errorf("genesis: no rom loaded")
	}

	snap := snapshot{
		CPU:          g.cpu.Snapshot(),
		Bus:          g.bus.Snapshot(),
		VDP:          g.video.Snapshot(),
		YM2612:       g.ym.Snapshot(),
		PSG:          g.sound.Snapshot(),
		Z80:          g.z80.Snapshot(),
		Z80Bus:       g.z80bus.Snapshot(),
		Z80Left:      g.z80Left,
		PrevZ80Reset: g.prevZ80Reset,
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("genesis: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (g *Genesis) LoadState(data []byte) error {
	if !g.loaded {
		return fmt.Errorf("genesis: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("genesis: load state: %w", err)
	}

	g.cpu.Restore(snap.CPU)
	g.bus.Restore(snap.Bus)
	g.video.Restore(snap.VDP)
	g.ym.Restore(snap.YM2612)
	g.sound.Restore(snap.PSG)
	g.z80.Restore(snap.Z80)
	g.z80bus.Restore(snap.Z80Bus)
	g.z80Left = snap.Z80Left
	g.prevZ80Reset = snap.PrevZ80Reset
	return nil
}
