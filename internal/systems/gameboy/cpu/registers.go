package cpu

// Flag bits within F, the low nibble of which is always zero.
const (
	FlagZ = 1 << 7 // Zero
	FlagN = 1 << 6 // Subtract
	FlagH = 1 << 5 // Half-carry
	FlagC = 1 << 4 // Carry
)

// registers holds the Sharp LR35902's eight 8-bit registers (paired as
// AF/BC/DE/HL), stack pointer and program counter.
type registers struct {
	A, F byte
	B, C byte
	D, E byte
	H, L byte
	SP   uint16
	PC   uint16
}

func (r *registers) AF() uint16 { return uint16(r.A)<<8 | uint16(r.F) }
func (r *registers) BC() uint16 { return uint16(r.B)<<8 | uint16(r.C) }
func (r *registers) DE() uint16 { return uint16(r.D)<<8 | uint16(r.E) }
func (r *registers) HL() uint16 { return uint16(r.H)<<8 | uint16(r.L) }

func (r *registers) SetAF(v uint16) { r.A, r.F = byte(v>>8), byte(v)&0xF0 }
func (r *registers) SetBC(v uint16) { r.B, r.C = byte(v>>8), byte(v) }
func (r *registers) SetDE(v uint16) { r.D, r.E = byte(v>>8), byte(v) }
func (r *registers) SetHL(v uint16) { r.H, r.L = byte(v>>8), byte(v) }

// SetFlag sets or clears the given flag bit(s) in F.
func (r *registers) SetFlag(mask byte, on bool) {
	if on {
		r.F |= mask
	} else {
		r.F &^= mask
	}
}

// HasFlag reports whether every bit in mask is set in F.
func (r *registers) HasFlag(mask byte) bool {
	return r.F&mask == mask
}
