package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// Button bit positions within the generic input.State bitmask,
// matching KEYINPUT's own bit order exactly so SetButtons is a
// straight copy.
const (
	A = iota
	B
	Select
	Start
	Right
	Left
	Up
	Down
	R
	L
)

// keypad tracks held buttons and produces KEYINPUT's active-low byte.
type keypad struct {
	buttons uint16
}

func (k *keypad) setButtons(state input.State) {
	var pressed uint16
	for bit := uint8(0); bit < 10; bit++ {
		if state.Pressed(bit) {
			pressed |= 1 << bit
		}
	}
	k.buttons = pressed
}

// read returns KEYINPUT: all 10 buttons active-low in the low bits,
// the rest reading as 1 (unused/always-released).
func (k *keypad) read() uint16 { return ^k.buttons & 0x03FF }
