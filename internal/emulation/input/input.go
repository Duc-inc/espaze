package input

// State is a generic 32-button bitmask shared by every core. Each core
// decides which bits it cares about (see its own input subpackage) and
// maps them to console-specific buttons/keys.
type State struct {
	Buttons uint32
}

// Pressed reports whether the given bit index (0-31) is currently held.
func (s State) Pressed(bit uint8) bool {
	if bit > 31 {
		return false
	}
	return s.Buttons&(1<<bit) != 0
}

// With returns a copy of the state with the given bit set to pressed/released.
func (s State) With(bit uint8, pressed bool) State {
	if bit > 31 {
		return s
	}
	if pressed {
		s.Buttons |= 1 << bit
	} else {
		s.Buttons &^= 1 << bit
	}
	return s
}
