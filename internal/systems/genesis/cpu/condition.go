package cpu

// evalCondition implements the 14 real branch/set/decrement conditions
// (codes 2-15; 0 and 1 are BRA/BSR or "always true/false" depending on
// context, handled by their own callers instead).
func (c *CPU) evalCondition(cc byte) bool {
	n, v, z, carry := c.regs.getFlag(FlagN), c.regs.getFlag(FlagV), c.regs.getFlag(FlagZ), c.regs.getFlag(FlagC)
	switch cc {
	case 0x0:
		return true
	case 0x1:
		return false
	case 0x2: // HI
		return !carry && !z
	case 0x3: // LS
		return carry || z
	case 0x4: // CC
		return !carry
	case 0x5: // CS
		return carry
	case 0x6: // NE
		return !z
	case 0x7: // EQ
		return z
	case 0x8: // VC
		return !v
	case 0x9: // VS
		return v
	case 0xA: // PL
		return !n
	case 0xB: // MI
		return n
	case 0xC: // GE
		return n == v
	case 0xD: // LT
		return n != v
	case 0xE: // GT
		return n == v && !z
	default: // LE
		return n != v || z
	}
}
