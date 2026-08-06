package input

// KeyMap translates frontend key identifiers (JS KeyboardEvent.code values,
// e.g. "Digit1", "KeyQ") into the generic State bit each one drives.
// Cores never see key names directly, only the resulting State.
type KeyMap map[string]uint8

// Resolve looks up the bit index bound to a frontend key code.
func (m KeyMap) Resolve(code string) (uint8, bool) {
	bit, ok := m[code]
	return bit, ok
}

// Apply folds a key event into an existing State, returning the updated one.
func (m KeyMap) Apply(state State, code string, pressed bool) State {
	bit, ok := m.Resolve(code)
	if !ok {
		return state
	}
	return state.With(bit, pressed)
}
