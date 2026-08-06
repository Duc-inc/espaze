package chip8

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/chip8/cpu"
	"github.com/Duc-inc/espaze/internal/systems/chip8/display"
	"github.com/Duc-inc/espaze/internal/systems/chip8/memory"
)

// snapshot captures every piece of state needed to resume emulation later:
// RAM, the screen, both timers and the CPU's registers/stack.
type snapshot struct {
	Memory  [memory.Size]byte
	Display [display.Width * display.Height]bool
	Delay   byte
	Sound   byte
	CPU     cpu.Snapshot
}

// SaveState implements core.Core.
func (c *Chip8) SaveState() ([]byte, error) {
	snap := snapshot{
		Memory:  c.mem.Snapshot(),
		Display: c.disp.Pixels(),
		Delay:   c.delay.Get(),
		Sound:   c.sound.Get(),
		CPU:     c.proc.Snapshot(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snap); err != nil {
		return nil, fmt.Errorf("chip8: save state: %w", err)
	}
	return buf.Bytes(), nil
}

// LoadState implements core.Core.
func (c *Chip8) LoadState(data []byte) error {
	var snap snapshot
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&snap); err != nil {
		return fmt.Errorf("chip8: load state: %w", err)
	}

	c.mem.Restore(snap.Memory)
	c.disp.Restore(snap.Display)
	c.delay.Set(snap.Delay)
	c.sound.Set(snap.Sound)
	c.proc.Restore(snap.CPU)
	return nil
}
