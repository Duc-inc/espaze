package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// Button bit positions within the generic input.State bitmask this
// project uses. Pause isn't part of either I/O port on real hardware -
// it's wired directly to the Z80's NMI line - so joypad.go only reads
// bits 0-5; bus.go handles bit 6 separately.
const (
	Up = iota
	Down
	Left
	Right
	Button1
	Button2
	Pause
)

// joypad tracks player 1's held buttons (this project only wires up a
// single player) and produces the two I/O ports' active-low bytes.
type joypad struct {
	state input.State
}

func (j *joypad) SetButtons(state input.State) { j.state = state }

// ReadPortDC implements $DC: P1's d-pad and two action buttons, plus
// the low 2 bits of P2's d-pad (always released, since P2 isn't wired
// up), all active-low.
func (j *joypad) ReadPortDC() byte {
	var v byte = 0xFF
	if j.state.Pressed(Up) {
		v &^= 1 << 0
	}
	if j.state.Pressed(Down) {
		v &^= 1 << 1
	}
	if j.state.Pressed(Left) {
		v &^= 1 << 2
	}
	if j.state.Pressed(Right) {
		v &^= 1 << 3
	}
	if j.state.Pressed(Button1) {
		v &^= 1 << 4
	}
	if j.state.Pressed(Button2) {
		v &^= 1 << 5
	}
	return v
}

// ReadPortDD implements $DD: the rest of P2's buttons (unused here) and
// two system bits this project always reports as "not asserted".
func (j *joypad) ReadPortDD() byte {
	return 0xFF
}
