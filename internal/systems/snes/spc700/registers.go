package spc700

// Processor status flag bits, in the SPC700's own documented positions.
const (
	FlagCarry     byte = 1 << 0
	FlagZero      byte = 1 << 1
	FlagIRQD      byte = 1 << 2
	FlagHalfCarry byte = 1 << 3
	FlagBreak     byte = 1 << 4
	FlagPage      byte = 1 << 5 // P: direct page base is $0100 instead of $0000 when set
	FlagOverflow  byte = 1 << 6
	FlagNegative  byte = 1 << 7
)

// registers holds the SPC700's real register set: three 8-bit
// accumulator/index registers (A/X/Y, with A+Y also usable as one
// 16-bit "YA" pair for a handful of instructions this project doesn't
// implement), an 8-bit stack pointer (SP is always $01xx - the SPC700
// has no separate stack page register), the flags register, and PC.
type registers struct {
	A, X, Y byte
	SP      byte
	PSW     byte
	PC      uint16
}

func (r *registers) getFlag(flag byte) bool { return r.PSW&flag != 0 }

func (r *registers) setFlag(flag byte, on bool) {
	if on {
		r.PSW |= flag
	} else {
		r.PSW &^= flag
	}
}

func (r *registers) setNZ(v byte) {
	r.setFlag(FlagZero, v == 0)
	r.setFlag(FlagNegative, v&0x80 != 0)
}

// directPageBase returns $0000 or $0100 depending on the P flag - the
// SPC700's own direct-page addressing quirk.
func (r *registers) directPageBase() uint16 {
	if r.getFlag(FlagPage) {
		return 0x0100
	}
	return 0x0000
}
