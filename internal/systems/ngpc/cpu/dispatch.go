package cpu

// dispatchEntry is one opcode byte's handler in this project's own
// TLCS900H-inspired instruction encoding (see this package's doc
// comment for why it's not the real chip's actual byte encoding).
type dispatchEntry struct {
	execute func(c *CPU) int
}

var dispatchTable [256]dispatchEntry

func setOp(opcode byte, fn func(c *CPU) int) {
	dispatchTable[opcode] = dispatchEntry{execute: fn}
}

// checkCondition evaluates one of this project's 7 condition codes.
func (c *CPU) checkCondition(cc byte) bool {
	switch cc {
	case 0:
		return c.regs.getFlag(FlagZero)
	case 1:
		return !c.regs.getFlag(FlagZero)
	case 2:
		return c.regs.getFlag(FlagCarry)
	case 3:
		return !c.regs.getFlag(FlagCarry)
	case 4:
		return c.regs.getFlag(FlagSign)
	case 5:
		return !c.regs.getFlag(FlagSign)
	case 6:
		return c.regs.getFlag(FlagOverflow)
	default:
		return true
	}
}

// hlAddr returns XHL's low 16 bits as an address - the "(HL)" indirect
// addressing mode every LD/ALU memory form in this encoding uses.
func (c *CPU) hlAddr() uint32 { return c.regs.XHL & 0xFFFF }
