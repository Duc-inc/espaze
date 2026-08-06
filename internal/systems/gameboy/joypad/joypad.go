package joypad

import "github.com/Duc-inc/espaze/internal/emulation/input"

// Button bit positions within both the generic input.State bitmask and
// Joypad's own internal "pressed" byte - kept identical so mapping is a
// straight copy.
const (
	Right = iota
	Left
	Up
	Down
	A
	B
	Select
	Start
)

// Joypad models the P1/JOYP register (0xFF00): games select either the
// four direction keys or the four button keys (or both) and read back
// which of the selected keys are currently held, active-low.
type Joypad struct {
	selectButtons   bool
	selectDirection bool
	pressed         uint8
}

// New returns a joypad with nothing selected and nothing pressed.
func New() *Joypad {
	return &Joypad{}
}

// Reset releases every key and clears the selection latch.
func (j *Joypad) Reset() {
	*j = Joypad{}
}

// SetButtons applies the latest generic input state, one bit per button
// using the same layout as the constants above.
func (j *Joypad) SetButtons(state input.State) {
	var pressed uint8
	for bit := uint8(0); bit < 8; bit++ {
		if state.Pressed(bit) {
			pressed |= 1 << bit
		}
	}
	j.pressed = pressed
}

// WriteRegister implements the CPU writing to 0xFF00 (only bits 4-5, the
// selection lines, are actually writable).
func (j *Joypad) WriteRegister(v byte) {
	j.selectDirection = v&0x10 == 0
	j.selectButtons = v&0x20 == 0
}

// ReadRegister implements the CPU reading 0xFF00.
func (j *Joypad) ReadRegister() byte {
	line := byte(0x0F)
	if j.selectDirection {
		line &^= directionBits(j.pressed)
	}
	if j.selectButtons {
		line &^= buttonBits(j.pressed)
	}

	result := byte(0xC0) | line
	if !j.selectDirection {
		result |= 0x10
	}
	if !j.selectButtons {
		result |= 0x20
	}
	return result
}

func directionBits(pressed uint8) byte {
	var out byte
	if pressed&(1<<Right) != 0 {
		out |= 0x01
	}
	if pressed&(1<<Left) != 0 {
		out |= 0x02
	}
	if pressed&(1<<Up) != 0 {
		out |= 0x04
	}
	if pressed&(1<<Down) != 0 {
		out |= 0x08
	}
	return out
}

func buttonBits(pressed uint8) byte {
	var out byte
	if pressed&(1<<A) != 0 {
		out |= 0x01
	}
	if pressed&(1<<B) != 0 {
		out |= 0x02
	}
	if pressed&(1<<Select) != 0 {
		out |= 0x04
	}
	if pressed&(1<<Start) != 0 {
		out |= 0x08
	}
	return out
}

// Snapshot/Restore persist the selection latch across save states (the
// live "pressed" bits are transient input and deliberately not restored).
type Snapshot struct {
	SelectButtons   bool
	SelectDirection bool
}

func (j *Joypad) Snapshot() Snapshot {
	return Snapshot{SelectButtons: j.selectButtons, SelectDirection: j.selectDirection}
}

func (j *Joypad) Restore(s Snapshot) {
	j.selectButtons = s.SelectButtons
	j.selectDirection = s.SelectDirection
}
