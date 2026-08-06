package input

// Keypad models the 16-key hex keypad (0-F) every CHIP-8 program expects.
type Keypad struct {
	keys [16]bool
}

// New returns a keypad with every key released.
func New() *Keypad {
	return &Keypad{}
}

// Set marks a key (0x0-0xF) as pressed or released; out-of-range keys are ignored.
func (k *Keypad) Set(key uint8, pressed bool) {
	if key > 0xF {
		return
	}
	k.keys[key] = pressed
}

// IsDown reports whether the given key is currently held.
func (k *Keypad) IsDown(key uint8) bool {
	if key > 0xF {
		return false
	}
	return k.keys[key]
}

// AnyDown returns the first key found pressed and true, or (0, false) if
// nothing is held. Used to implement the "wait for key" (Fx0A) instruction.
func (k *Keypad) AnyDown() (uint8, bool) {
	for i, down := range k.keys {
		if down {
			return uint8(i), true
		}
	}
	return 0, false
}

// Reset releases every key.
func (k *Keypad) Reset() {
	k.keys = [16]bool{}
}
