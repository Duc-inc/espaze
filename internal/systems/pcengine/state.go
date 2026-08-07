package pcengine

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/pcengine/cpu"
	"github.com/Duc-inc/espaze/internal/systems/pcengine/memory"
	"github.com/Duc-inc/espaze/internal/systems/pcengine/psg"
	"github.com/Duc-inc/espaze/internal/systems/pcengine/vce"
	"github.com/Duc-inc/espaze/internal/systems/pcengine/vdc"
)

// snapshot combines every component's own snapshot into one save
// state. The cartridge ROM itself is never included - LoadState
// assumes the same ROM is already loaded, exactly like every other
// core here.
type snapshot struct {
	CPU cpu.Snapshot
	Bus memory.Snapshot
	VDC vdc.Snapshot
	VCE vce.Snapshot
	PSG psg.Snapshot
}

// SaveState implements core.Core.
func (p *PCEngine) SaveState() ([]byte, error) {
	if !p.loaded {
		return nil, fmt.Errorf("pcengine: no rom loaded")
	}

	snap := snapshot{
		CPU: p.proc.Snapshot(),
		Bus: p.bus.Snapshot(),
		VDC: p.video.Snapshot(),
		VCE: p.color.Snapshot(),
		PSG: p.sound.Snapshot(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("pcengine: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (p *PCEngine) LoadState(data []byte) error {
	if !p.loaded {
		return fmt.Errorf("pcengine: no rom loaded")
	}

	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("pcengine: load state: %w", err)
	}

	p.proc.Restore(snap.CPU)
	p.bus.Restore(snap.Bus)
	p.video.Restore(snap.VDC)
	p.color.Restore(snap.VCE)
	p.sound.Restore(snap.PSG)
	return nil
}
