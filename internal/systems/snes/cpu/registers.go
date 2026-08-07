package cpu

// Processor status flag bits.
const (
	FlagCarry    byte = 1 << 0
	FlagZero     byte = 1 << 1
	FlagIRQD     byte = 1 << 2
	FlagDecimal  byte = 1 << 3
	FlagIndex8   byte = 1 << 4 // X: 1 = 8-bit X/Y (native mode only)
	FlagAccum8   byte = 1 << 5 // M: 1 = 8-bit A (native mode only)
	FlagBreak    byte = 1 << 4 // same bit as FlagIndex8, meaning depends on E
	FlagOverflow byte = 1 << 6
	FlagNegative byte = 1 << 7
)

// registers holds the 65816's visible register set: a 16-bit
// accumulator/index registers (each independently switchable to 8-bit
// via the M/X status flags), the Direct Page, Data Bank, and Program
// Bank registers that give it a 24-bit address space despite a 16-bit
// program counter, and the emulation-mode latch (E) that isn't part
// of P but controls how several instructions and the stack behave -
// exactly matching real hardware's own quirky "P holds most flags,
// but E lives outside it" design.
type registers struct {
	A, X, Y uint16
	D       uint16
	S       uint16
	PC      uint16
	DBR     byte
	PBR     byte
	P       byte
	E       bool // emulation mode
}

func (r *registers) getFlag(flag byte) bool { return r.P&flag != 0 }

func (r *registers) setFlag(flag byte, on bool) {
	if on {
		r.P |= flag
	} else {
		r.P &^= flag
	}
}

// accum8/index8 report whether A or X/Y are currently in 8-bit mode -
// always true in emulation mode, otherwise governed by the M/X flags.
func (r *registers) accum8() bool { return r.E || r.getFlag(FlagAccum8) }
func (r *registers) index8() bool { return r.E || r.getFlag(FlagIndex8) }

func (r *registers) setNZ8(v byte) {
	r.setFlag(FlagZero, v == 0)
	r.setFlag(FlagNegative, v&0x80 != 0)
}

func (r *registers) setNZ16(v uint16) {
	r.setFlag(FlagZero, v == 0)
	r.setFlag(FlagNegative, v&0x8000 != 0)
}
