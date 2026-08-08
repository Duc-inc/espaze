package powerpc

func init() {
	setPrimary(24, func(c *CPU, instr uint32) int { // ori
		rS, rA := fieldRD(instr), fieldRA(instr)
		c.regs.GPR[rA] = c.regs.GPR[rS] | fieldUimm(instr)
		return 2
	})
	setPrimary(25, func(c *CPU, instr uint32) int { // oris
		rS, rA := fieldRD(instr), fieldRA(instr)
		c.regs.GPR[rA] = c.regs.GPR[rS] | fieldUimm(instr)<<16
		return 2
	})
	setPrimary(26, func(c *CPU, instr uint32) int { // xori
		rS, rA := fieldRD(instr), fieldRA(instr)
		c.regs.GPR[rA] = c.regs.GPR[rS] ^ fieldUimm(instr)
		return 2
	})
	setPrimary(27, func(c *CPU, instr uint32) int { // xoris
		rS, rA := fieldRD(instr), fieldRA(instr)
		c.regs.GPR[rA] = c.regs.GPR[rS] ^ fieldUimm(instr)<<16
		return 2
	})
	setPrimary(28, func(c *CPU, instr uint32) int { // andi.
		rS, rA := fieldRD(instr), fieldRA(instr)
		result := c.regs.GPR[rS] & fieldUimm(instr)
		c.regs.GPR[rA] = result
		c.regs.setCR0(result)
		return 2
	})
	setPrimary(29, func(c *CPU, instr uint32) int { // andis.
		rS, rA := fieldRD(instr), fieldRA(instr)
		result := c.regs.GPR[rS] & fieldUimm(instr) << 16
		c.regs.GPR[rA] = result
		c.regs.setCR0(result)
		return 2
	})

	setExt31(28, func(c *CPU, instr uint32) int { // and
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := c.regs.GPR[rS] & c.regs.GPR[rB]
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(444, func(c *CPU, instr uint32) int { // or
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := c.regs.GPR[rS] | c.regs.GPR[rB]
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(316, func(c *CPU, instr uint32) int { // xor
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := c.regs.GPR[rS] ^ c.regs.GPR[rB]
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(476, func(c *CPU, instr uint32) int { // nand
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := ^(c.regs.GPR[rS] & c.regs.GPR[rB])
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(124, func(c *CPU, instr uint32) int { // nor
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := ^(c.regs.GPR[rS] | c.regs.GPR[rB])
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(284, func(c *CPU, instr uint32) int { // eqv
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := ^(c.regs.GPR[rS] ^ c.regs.GPR[rB])
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(60, func(c *CPU, instr uint32) int { // andc
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := c.regs.GPR[rS] &^ c.regs.GPR[rB]
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(412, func(c *CPU, instr uint32) int { // orc
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := c.regs.GPR[rS] | ^c.regs.GPR[rB]
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(954, func(c *CPU, instr uint32) int { // extsb
		rS, rA := fieldRD(instr), fieldRA(instr)
		result := uint32(int32(int8(c.regs.GPR[rS])))
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(922, func(c *CPU, instr uint32) int { // extsh
		rS, rA := fieldRD(instr), fieldRA(instr)
		result := uint32(int32(int16(c.regs.GPR[rS])))
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
}
