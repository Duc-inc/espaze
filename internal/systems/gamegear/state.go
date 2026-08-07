package gamegear

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/sms/cpu"
	"github.com/Duc-inc/espaze/internal/systems/sms/memory"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
	"github.com/Duc-inc/espaze/internal/systems/sms/vdp"
)

// snapshot combines every reused component's own snapshot into one
// save state, plus the ioBus wrapper's own Start-button latch. The
// cartridge ROM itself is never included - LoadState assumes the same
// ROM is already loaded, exactly like every other core here.
type snapshot struct {
	CPU   cpu.Snapshot
	Bus   memory.Snapshot
	VDP   vdp.Snapshot
	PSG   psg.Snapshot
	Start bool
}

// SaveState implements core.Core.
func (g *GameGear) SaveState() ([]byte, error) {
	if !g.loaded {
		return nil, fmt.Errorf("gamegear: no rom loaded")
	}

	snap := snapshot{
		CPU:   g.proc.Snapshot(),
		Bus:   g.bus.Snapshot(),
		VDP:   g.video.Snapshot(),
		PSG:   g.sound.Snapshot(),
		Start: g.io.start,
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("gamegear: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (g *GameGear) LoadState(data []byte) error {
	if !g.loaded {
		return fmt.Errorf("gamegear: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("gamegear: load state: %w", err)
	}

	g.proc.Restore(snap.CPU)
	g.bus.Restore(snap.Bus)
	g.video.Restore(snap.VDP)
	g.sound.Restore(snap.PSG)
	g.io.start = snap.Start
	return nil
}
