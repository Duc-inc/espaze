package memory

import "github.com/Duc-inc/espaze/internal/emulation/input"

// Button bit positions within the generic input.State bitmask.
const (
	Up = iota
	Down
	Left
	Right
	A
	B
	Option
)

// Controller tracks held buttons and produces this project's active-
// low input port byte.
type Controller struct {
	buttons byte
}

// SetButtons applies the latest generic input state.
func (c *Controller) SetButtons(state input.State) {
	var pressed byte
	for bit := uint8(0); bit < 7; bit++ {
		if state.Pressed(bit) {
			pressed |= 1 << bit
		}
	}
	c.buttons = pressed
}

// Read returns the input port: all 7 buttons active-low in the low
// bits, the rest reading as 1.
func (c *Controller) Read() byte { return ^c.buttons }
