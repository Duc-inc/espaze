package colecovision

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/colecovision/memory"
	"github.com/Duc-inc/espaze/internal/systems/sms/cpu"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
	"github.com/Duc-inc/espaze/internal/systems/tms9918"
)

// snapshot combines every component's own snapshot into one save
// state. The cartridge ROM itself is never included - LoadState
// assumes the same ROM is already loaded, exactly like every other
// core here.
type snapshot struct {
	CPU cpu.Snapshot
	Bus memory.Snapshot
	VDP tms9918.Snapshot
	PSG psg.Snapshot
}

// SaveState implements core.Core.
func (c *ColecoVision) SaveState() ([]byte, error) {
	if !c.loaded {
		return nil, fmt.Errorf("colecovision: no rom loaded")
	}

	snap := snapshot{
		CPU: c.proc.Snapshot(), Bus: c.bus.Snapshot(),
		VDP: c.video.Snapshot(), PSG: c.sound.Snapshot(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("colecovision: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (c *ColecoVision) LoadState(data []byte) error {
	if !c.loaded {
		return fmt.Errorf("colecovision: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("colecovision: load state: %w", err)
	}

	c.proc.Restore(snap.CPU)
	c.bus.Restore(snap.Bus)
	c.video.Restore(snap.VDP)
	c.sound.Restore(snap.PSG)
	return nil
}
