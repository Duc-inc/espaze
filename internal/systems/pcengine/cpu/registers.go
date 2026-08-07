package cpu

// Status flag bits, in the same bit positions as every 6502-family chip.
const (
	FlagCarry     byte = 1 << 0
	FlagZero      byte = 1 << 1
	FlagInterrupt byte = 1 << 2
	FlagDecimal   byte = 1 << 3
	FlagBreak     byte = 1 << 4
	FlagUnused    byte = 1 << 5
	FlagOverflow  byte = 1 << 6
	FlagNegative  byte = 1 << 7
)

// registers holds the HuC6280's visible register set - the same
// layout as any 6502-family chip (no extra user-visible registers;
// the memory-mapping unit's 8 page registers live in mmu.go instead).
type registers struct {
	A, X, Y byte
	S       byte
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

func (r *registers) getFlag(flag byte) bool { return r.P&flag != 0 }

func (r *registers) setNZ(v byte) {
	r.setFlag(FlagZero, v == 0)
	r.setFlag(FlagNegative, v&0x80 != 0)
}
