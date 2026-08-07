package powerpc

func (c *CPU) gprOrZero(r uint32) uint32 {
	if r == 0 {
		return 0
	}
	return c.regs.GPR[r]
}

func init() {
	setPrimary(14, func(c *CPU, instr uint32) int { // addi
		rD, rA := fieldRD(instr), fieldRA(instr)
		c.regs.GPR[rD] = c.gprOrZero(rA) + uint32(fieldSimm(instr))
		return 2
	})
	setPrimary(15, func(c *CPU, instr uint32) int { // addis
		rD, rA := fieldRD(instr), fieldRA(instr)
		c.regs.GPR[rD] = c.gprOrZero(rA) + uint32(fieldSimm(instr))<<16
		return 2
	})
	setPrimary(7, func(c *CPU, instr uint32) int { // mulli
		rD, rA := fieldRD(instr), fieldRA(instr)
		c.regs.GPR[rD] = uint32(int32(c.regs.GPR[rA]) * fieldSimm(instr))
		return 3
	})

	setExt31(266, func(c *CPU, instr uint32) int { // add
		rD, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := c.regs.GPR[rA] + c.regs.GPR[rB]
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(40, func(c *CPU, instr uint32) int { // subf
		rD, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := c.regs.GPR[rB] - c.regs.GPR[rA]
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(235, func(c *CPU, instr uint32) int { // mullw
		rD, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := uint32(int32(c.regs.GPR[rA]) * int32(c.regs.GPR[rB]))
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 4
	})
	setExt31(491, func(c *CPU, instr uint32) int { // divw
		rD, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		if c.regs.GPR[rB] == 0 {
			return 4
		}
		result := uint32(int32(c.regs.GPR[rA]) / int32(c.regs.GPR[rB]))
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 20
	})
}
