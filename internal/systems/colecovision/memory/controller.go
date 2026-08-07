package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// Button bit positions within the generic input.State bitmask.
const (
	Up = iota
	Down
	Left
	Right
	Fire1
	Fire2
)

// Controller is a simplified ColecoVision pad: joystick + two fire
// buttons only. Real hardware also multiplexes a 12-key numeric
// keypad through the same port (selected by writing to a separate
// mode-select port before reading) - this project always reports "no
// key held" for that, so games that require keypad entry (a
// documented handful, mostly for save codes or game-select menus)
// won't be usable through it.
type Controller struct {
	buttons byte
}

// SetButtons applies the latest generic input state.
func (c *Controller) SetButtons(state input.State) {
	var pressed byte
	for bit := uint8(0); bit < 6; bit++ {
		if state.Pressed(bit) {
			pressed |= 1 << bit
		}
	}
	c.buttons = pressed
}

// Read returns the joystick port's active-low byte.
func (c *Controller) Read() byte {
	var v byte = 0xFF
	if c.buttons&(1<<Up) != 0 {
		v &^= 1 << 0
	}
	if c.buttons&(1<<Down) != 0 {
		v &^= 1 << 1
	}
	if c.buttons&(1<<Left) != 0 {
		v &^= 1 << 2
	}
	if c.buttons&(1<<Right) != 0 {
		v &^= 1 << 3
	}
	if c.buttons&(1<<Fire1) != 0 {
		v &^= 1 << 6
	}
	if c.buttons&(1<<Fire2) != 0 {
		v &^= 1 << 5
	}
	return v
}
