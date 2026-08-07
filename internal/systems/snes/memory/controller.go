package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// Button bit positions within the generic input.State bitmask.
const (
	B = iota
	Y
	Select
	Start
	Up
	Down
	Left
	Right
	A
	X
	L
	R
)

// Controller tracks held buttons and produces a 16-bit active-high
// bitfield (matching the real SNES pad's own report format, unlike
// most of this project's other active-low controllers).
type Controller struct {
	buttons uint16
}

// SetButtons applies the latest generic input state.
func (c *Controller) SetButtons(state input.State) {
	var pressed uint16
	for bit := uint8(0); bit < 12; bit++ {
		if state.Pressed(bit) {
			pressed |= 1 << bit
		}
	}
	c.buttons = pressed
}

// ReadLow/ReadHigh return the two bytes of the controller's report.
func (c *Controller) ReadLow() byte  { return byte(c.buttons) }
func (c *Controller) ReadHigh() byte { return byte(c.buttons >> 8) }
