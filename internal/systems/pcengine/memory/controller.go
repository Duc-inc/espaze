package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// Button bit positions within the generic input.State bitmask.
const (
	Up = iota
	Down
	Left
	Right
	ButtonI
	ButtonII
	Select
	Run
)

// Controller is a standard PC Engine 2-button pad: writing the SEL
// line to the I/O port picks whether the next read returns the d-pad
// or the four face/system buttons, both active-low, matching this
// project's other cores' convention (this project hasn't
// independently confirmed active-low is exactly right for the real
// PCE pad, but it's the commonly documented behavior).
type Controller struct {
	sel     bool
	buttons byte
}

// SetButtons applies the latest generic input state.
func (c *Controller) SetButtons(state input.State) {
	var pressed byte
	for bit := uint8(0); bit < 8; bit++ {
		if state.Pressed(bit) {
			pressed |= 1 << bit
		}
	}
	c.buttons = pressed
}

// Write latches the SEL line from an I/O port write.
func (c *Controller) Write(v byte) { c.sel = v&0x01 != 0 }

// Read returns the currently selected half of the pad's state.
func (c *Controller) Read() byte {
	if c.sel {
		var v byte = 0x0F
		if c.buttons&(1<<Up) != 0 {
			v &^= 1 << 0
		}
		if c.buttons&(1<<Right) != 0 {
			v &^= 1 << 1
		}
		if c.buttons&(1<<Down) != 0 {
			v &^= 1 << 2
		}
		if c.buttons&(1<<Left) != 0 {
			v &^= 1 << 3
		}
		return v | 0xF0
	}
	var v byte = 0x0F
	if c.buttons&(1<<ButtonI) != 0 {
		v &^= 1 << 0
	}
	if c.buttons&(1<<ButtonII) != 0 {
		v &^= 1 << 1
	}
	if c.buttons&(1<<Select) != 0 {
		v &^= 1 << 2
	}
	if c.buttons&(1<<Run) != 0 {
		v &^= 1 << 3
	}
	return v | 0xF0
}
