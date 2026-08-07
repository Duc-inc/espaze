package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// Button bit positions within the generic input.State bitmask.
const (
	Up = iota
	Down
	Left
	Right
	B
	C
	A
	Start
)

// Controller is a standard 3-button Genesis pad, wired to port A only
// - port B and the EXT port aren't implemented (single-controller,
// matching this project's other cores). Real hardware multiplexes Up/
// Down/B/C onto the low bits when TH is driven high and Up/Down/A/
// Start when driven low; software toggles TH via the data port itself
// (the direction register is always assumed to configure it as an
// output, which is what every game actually does).
type Controller struct {
	buttons byte
	th      bool
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

func (c *Controller) pressed(bit int) bool { return c.buttons&(1<<uint(bit)) != 0 }

func low(pressed bool) byte {
	if pressed {
		return 0
	}
	return 1
}

// Read returns the data port's current byte, matching the pad's
// active-low button lines.
func (c *Controller) Read() byte {
	base := byte(0)
	if c.th {
		base = 0x40
	}
	if c.th {
		return base | low(c.pressed(C))<<5 | low(c.pressed(B))<<4 |
			low(c.pressed(Right))<<3 | low(c.pressed(Left))<<2 |
			low(c.pressed(Down))<<1 | low(c.pressed(Up))
	}
	return base | low(c.pressed(Start))<<5 | low(c.pressed(A))<<4 | 0x0C |
		low(c.pressed(Down))<<1 | low(c.pressed(Up))
}

// Write latches the TH output line from a data port write.
func (c *Controller) Write(v byte) { c.th = v&0x40 != 0 }
