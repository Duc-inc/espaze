package gamegear

import "github.com/Duc-inc/espaze/internal/systems/sms/memory"

// ioBus wraps the SMS bus's I/O port space with the one thing that's
// actually different about the Game Gear: a dedicated Start button and
// region byte at port $00 (the SMS has no such port; its own Pause
// button is wired to NMI instead, which the Game Gear doesn't have -
// so Reset here never fires PausePressed). Every other port ($40 and
// up: V/H counters, VDP, PSG, joypad) is identical hardware, so those
// just delegate straight through.
type ioBus struct {
	mem   *memory.Bus
	start bool
}

// In implements sms/cpu.IOBus.
func (b *ioBus) In(port byte) byte {
	if port == 0x00 {
		v := byte(0x40) // bit6 always reads 0 (region: export); bit7 is Start, active-low
		if !b.start {
			v |= 0x80
		}
		return v
	}
	return b.mem.In(port)
}

// Out implements sms/cpu.IOBus. Port $00 is read-only on real hardware.
func (b *ioBus) Out(port byte, v byte) {
	if port == 0x00 {
		return
	}
	b.mem.Out(port, v)
}

// SetStart applies the Start button's current state.
func (b *ioBus) SetStart(pressed bool) { b.start = pressed }
