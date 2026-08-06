package cpu

import "github.com/Duc-inc/espaze/internal/systems/chip8/memory"

const programStart = memory.ProgramStart

// registers holds the general-purpose and special-purpose CPU registers,
// identical in shape to base CHIP-8's.
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
