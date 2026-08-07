package powerpc

// setCRField writes a signed/unsigned-compare 4-bit result into CR
// field 0 - this project always targets field 0, simplifying away the
// 3-bit crfD selector real compare instructions have for CR1-CR7.
func (c *CPU) setCRField(lt, gt, eq bool) {
	var field uint32
	switch {
	case lt:
		field = 0x8
	case gt:
		field = 0x4
	case eq:
		field = 0x2
	}
	c.regs.CR = c.regs.CR&0x0FFFFFFF | field<<28
}

func init() {
	setPrimary(11, func(c *CPU, instr uint32) int { // cmpi
		rA := fieldRA(instr)
		a, b := int32(c.regs.GPR[rA]), fieldSimm(instr)
		c.setCRField(a < b, a > b, a == b)
		return 2
	})
	setPrimary(10, func(c *CPU, instr uint32) int { // cmpli
		rA := fieldRA(instr)
		a, b := c.regs.GPR[rA], fieldUimm(instr)
		c.setCRField(a < b, a > b, a == b)
		return 2
	})

	setExt31(0, func(c *CPU, instr uint32) int { // cmp
		rA, rB := fieldRA(instr), fieldRB(instr)
		a, b := int32(c.regs.GPR[rA]), int32(c.regs.GPR[rB])
		c.setCRField(a < b, a > b, a == b)
		return 2
	})
	setExt31(32, func(c *CPU, instr uint32) int { // cmpl
		rA, rB := fieldRA(instr), fieldRB(instr)
		a, b := c.regs.GPR[rA], c.regs.GPR[rB]
		c.setCRField(a < b, a > b, a == b)
		return 2
	})
}
