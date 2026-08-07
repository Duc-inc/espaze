package cpu

// dispatchEntry is one opcode byte's handler; each resolves its own
// addressing mode and returns an approximate cycle cost.
type dispatchEntry struct {
	execute func(c *CPU) int
}

var dispatchTable [256]dispatchEntry

func setOp(opcode byte, fn func(c *CPU) int) {
	dispatchTable[opcode] = dispatchEntry{execute: fn}
}

func (c *CPU) checkCondition(cc byte) bool {
	switch cc {
	case 0:
		return !c.regs.getFlag(FlagNegative) // BPL
	case 1:
		return c.regs.getFlag(FlagNegative) // BMI
	case 2:
		return !c.regs.getFlag(FlagOverflow) // BVC
	case 3:
		return c.regs.getFlag(FlagOverflow) // BVS
	case 4:
		return !c.regs.getFlag(FlagCarry) // BCC
	case 5:
		return c.regs.getFlag(FlagCarry) // BCS
	case 6:
		return !c.regs.getFlag(FlagZero) // BNE
	default:
		return c.regs.getFlag(FlagZero) // BEQ
	}
}
