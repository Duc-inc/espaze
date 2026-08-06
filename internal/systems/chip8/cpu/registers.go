package cpu

// registers holds the general-purpose and special-purpose CPU registers.
// V0-VE are general purpose, VF doubles as a flag register set by several
// instructions (carry, borrow, collision). I is the address register.
type registers struct {
	V  [16]byte
	I  uint16
	PC uint16
}

func newRegisters() registers {
	return registers{PC: programStart}
}

func (r *registers) reset() {
	*r = newRegisters()
}
