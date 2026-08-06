package cpu

// Status flag bits within the P register.
const (
	FlagCarry     byte = 1 << 0
	FlagZero      byte = 1 << 1
	FlagInterrupt byte = 1 << 2
	FlagDecimal   byte = 1 << 3 // present on real hardware, has no effect on the NES's 2A03
	FlagBreak     byte = 1 << 4
	FlagUnused    byte = 1 << 5 // always reads back as 1
	FlagOverflow  byte = 1 << 6
	FlagNegative  byte = 1 << 7
)

// registers holds the 6502's whole visible state: three 8-bit general
// registers, the stack pointer, the program counter, and the status
// flags packed into one byte.
type registers struct {
	A, X, Y byte
	SP      byte
	PC      uint16
	P       byte
}

func (r *registers) setFlag(flag byte, on bool) {
	if on {
		r.P |= flag
	} else {
		r.P &^= flag
	}
}

func (r *registers) getFlag(flag byte) bool {
	return r.P&flag != 0
}

// setZN sets the Zero and Negative flags from a just-computed byte
// result, the pattern nearly every instruction that touches memory or a
// register ends with.
func (r *registers) setZN(v byte) {
	r.setFlag(FlagZero, v == 0)
	r.setFlag(FlagNegative, v&0x80 != 0)
}
