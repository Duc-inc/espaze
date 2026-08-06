package schip

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/chip8/memory"
	"github.com/Duc-inc/espaze/internal/systems/schip/cpu"
)

// snapshot captures every piece of state needed to resume emulation later:
// RAM, the screen (including its current resolution mode), both timers
// and the CPU's registers/stack/RPL flags.
type snapshot struct {
	Memory          [memory.Size]byte
	DisplayExtended bool
	DisplayWidth    int
	DisplayHeight   int
	DisplayPixels   []bool
	Delay           byte
	Sound           byte
	CPU             cpu.Snapshot
}

// SaveState implements core.Core.
func (s *Schip) SaveState() ([]byte, error) {
	snap := snapshot{
		Memory:          s.mem.Snapshot(),
		DisplayExtended: s.disp.Extended(),
		DisplayWidth:    s.disp.Width(),
		DisplayHeight:   s.disp.Height(),
		DisplayPixels:   s.disp.Pixels(),
		Delay:           s.delay.Get(),
		Sound:           s.sound.Get(),
		CPU:             s.proc.Snapshot(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("schip: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (s *Schip) LoadState(data []byte) error {
	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("schip: load state: %w", err)
	}

	s.mem.Restore(snap.Memory)
	s.disp.Restore(snap.DisplayExtended, snap.DisplayWidth, snap.DisplayHeight, snap.DisplayPixels)
	s.delay.Set(snap.Delay)
	s.sound.Set(snap.Sound)
	s.proc.Restore(snap.CPU)
	return nil
}
