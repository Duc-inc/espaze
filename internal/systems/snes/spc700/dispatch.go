package spc700

type dispatchEntry struct {
	execute func(c *CPU) int
}

var dispatchTable [256]dispatchEntry

func setOp(opcode byte, fn func(c *CPU) int) {
	dispatchTable[opcode] = dispatchEntry{execute: fn}
}

func (c *CPU) directPage() uint16 { return c.regs.directPageBase() + uint16(c.fetch8()) }

func (c *CPU) checkCondition(cc byte) bool {
	switch cc {
	case 0:
		return c.regs.getFlag(FlagZero) // BEQ
	case 1:
		return !c.regs.getFlag(FlagZero) // BNE
	case 2:
		return c.regs.getFlag(FlagCarry) // BCS
	case 3:
		return !c.regs.getFlag(FlagCarry) // BCC
	case 4:
		return c.regs.getFlag(FlagNegative) // BMI
	case 5:
		return !c.regs.getFlag(FlagNegative) // BPL
	case 6:
		return c.regs.getFlag(FlagOverflow) // BVS
	default:
		return true
	}
}

// regPtr returns a pointer to A/X/Y by this project's own 0/1/2 coding.
func (c *CPU) regPtr(code byte) *byte {
	switch code & 0x03 {
	case 1:
		return &c.regs.X
	case 2:
		return &c.regs.Y
	default:
		return &c.regs.A
	}
}
