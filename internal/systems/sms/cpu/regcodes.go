package cpu

// The Z80's opcode byte splits into bit fields x(7-6) y(5-3) z(2-0),
// with p=y>>1 and q=y&1 - the standard decomposition (see Sean Young's
// "The Undocumented Z80 Documented") that makes its main opcode space
// regular enough to decode by table lookup on these fields instead of
// one throwaway case per opcode.

func decomposeOpcode(op byte) (x, y, z, p, q byte) {
	x = op >> 6
	y = (op >> 3) & 0x07
	z = op & 0x07
	p = y >> 1
	q = y & 1
	return
}

// r8 reads one of the 8 single-register operands the z/y fields select
// (index 6 means "(HL)", handled by the caller since it needs a memory
// access instead of a register read).
func (c *CPU) r8(index byte) byte {
	switch index {
	case 0:
		return c.regs.B
	case 1:
		return c.regs.C
	case 2:
		return c.regs.D
	case 3:
		return c.regs.E
	case 4:
		return c.regs.H
	case 5:
		return c.regs.L
	case 6:
		return c.bus.Read(c.regs.HL())
	default:
		return c.regs.A
	}
}

func (c *CPU) setR8(index byte, v byte) {
	switch index {
	case 0:
		c.regs.B = v
	case 1:
		c.regs.C = v
	case 2:
		c.regs.D = v
	case 3:
		c.regs.E = v
	case 4:
		c.regs.H = v
	case 5:
		c.regs.L = v
	case 6:
		c.bus.Write(c.regs.HL(), v)
	default:
		c.regs.A = v
	}
}

// rp reads one of the 4 register-pair operands the p field selects in
// most contexts (16-bit loads, INC/DEC rp, ADD HL,rp).
func (c *CPU) rp(index byte) uint16 {
	switch index {
	case 0:
		return c.regs.BC()
	case 1:
		return c.regs.DE()
	case 2:
		return c.regs.HL()
	default:
		return c.regs.SP
	}
}

func (c *CPU) setRP(index byte, v uint16) {
	switch index {
	case 0:
		c.regs.SetBC(v)
	case 1:
		c.regs.SetDE(v)
	case 2:
		c.regs.SetHL(v)
	default:
		c.regs.SP = v
	}
}

// rp2 is rp's sibling for PUSH/POP, where slot 3 is AF instead of SP.
func (c *CPU) rp2(index byte) uint16 {
	if index == 3 {
		return c.regs.AF()
	}
	return c.rp(index)
}

func (c *CPU) setRP2(index byte, v uint16) {
	if index == 3 {
		c.regs.SetAF(v)
		return
	}
	c.setRP(index, v)
}

// condition evaluates one of the 8 branch conditions the y field
// selects for JR/JP/CALL/RET.
func (c *CPU) condition(index byte) bool {
	switch index {
	case 0:
		return !c.regs.getFlag(FlagZ)
	case 1:
		return c.regs.getFlag(FlagZ)
	case 2:
		return !c.regs.getFlag(FlagC)
	case 3:
		return c.regs.getFlag(FlagC)
	case 4:
		return !c.regs.getFlag(FlagPV)
	case 5:
		return c.regs.getFlag(FlagPV)
	case 6:
		return !c.regs.getFlag(FlagS)
	default:
		return c.regs.getFlag(FlagS)
	}
}
