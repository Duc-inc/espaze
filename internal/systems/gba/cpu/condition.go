package cpu

// checkCondition evaluates one of ARM's 16 4-bit condition codes
// against the current flags - used both by ARM's per-instruction
// condition field and Thumb's conditional branch instruction.
func (c *CPU) checkCondition(cond uint32) bool {
	n := c.regs.getFlag(FlagN)
	z := c.regs.getFlag(FlagZ)
	cf := c.regs.getFlag(FlagC)
	v := c.regs.getFlag(FlagV)

	switch cond {
	case 0x0:
		return z // EQ
	case 0x1:
		return !z // NE
	case 0x2:
		return cf // CS/HS
	case 0x3:
		return !cf // CC/LO
	case 0x4:
		return n // MI
	case 0x5:
		return !n // PL
	case 0x6:
		return v // VS
	case 0x7:
		return !v // VC
	case 0x8:
		return cf && !z // HI
	case 0x9:
		return !cf || z // LS
	case 0xA:
		return n == v // GE
	case 0xB:
		return n != v // LT
	case 0xC:
		return !z && n == v // GT
	case 0xD:
		return z || n != v // LE
	case 0xE:
		return true // AL
	default:
		return false // reserved (0xF, "NV")
	}
}
