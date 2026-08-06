package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// Button bit positions within both the generic input.State bitmask and
// Controller's own internal byte - kept identical so mapping is a
// straight copy, and matching the real hardware's shift-out order.
const (
	A = iota
	B
	Select
	Start
	Up
	Down
	Left
	Right
)

// Controller is a standard NES gamepad: on a $4016 write with bit0 set,
// it latches the current button state; while strobed low, each $4016/17
// read shifts out one button at a time in the order above, matching the
// real controller's shift register.
type Controller struct {
	buttons byte
	shift   byte
	strobe  bool
}

// SetButtons applies the latest generic input state, one bit per button
// using the same layout as the constants above.
func (c *Controller) SetButtons(state input.State) {
	var pressed byte
	for bit := uint8(0); bit < 8; bit++ {
		if state.Pressed(bit) {
			pressed |= 1 << bit
		}
	}
	c.buttons = pressed
}

// Write handles a CPU write to $4016 (the strobe line both controllers share).
func (c *Controller) Write(v byte) {
	c.strobe = v&1 != 0
	if c.strobe {
		c.shift = c.buttons
	}
}

// Read handles a CPU read of this controller's port ($4016 or $4017).
// Real hardware keeps shifting out 1s once all 8 buttons are read; open
// bus bits above bit0 are approximated as 0 here.
func (c *Controller) Read() byte {
	if c.strobe {
		return c.buttons & 1
	}
	bit := c.shift & 1
	c.shift = c.shift>>1 | 0x80
	return bit
}
