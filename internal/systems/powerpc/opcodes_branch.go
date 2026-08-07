package powerpc

// branchTaken evaluates the BO/BI condition fields shared by every
// conditional branch form (bc/bclr/bcctr) - PowerPC's own documented
// semantics: BO's top bit skips the CTR decrement-and-test, its next
// bit picks which CTR result branches, its third bit skips the CR
// bit test entirely, and its fourth bit picks which CR bit value
// branches.
func (c *CPU) branchTaken(instr uint32) bool {
	bo := fieldBO(instr)

	ctrCond := true
	if bo&0x10 == 0 {
		c.regs.CTR--
		wantNotZero := bo&0x08 == 0
		ctrCond = (c.regs.CTR != 0) == wantNotZero
	}

	crCond := true
	if bo&0x04 == 0 {
		bi := fieldBI(instr)
		bitSet := c.regs.CR&(0x80000000>>bi) != 0
		wantSet := bo&0x02 != 0
		crCond = bitSet == wantSet
	}

	return ctrCond && crCond
}

func init() {
	setPrimary(18, func(c *CPU, instr uint32) int { // b/ba/bl/bla
		here := c.regs.PC - 4
		target := uint32(int32(here) + fieldLI(instr))
		if fieldAA(instr) {
			target = uint32(fieldLI(instr))
		}
		if fieldLK(instr) {
			c.regs.LR = c.regs.PC
		}
		c.regs.PC = target
		return 4
	})

	setPrimary(16, func(c *CPU, instr uint32) int { // bc/bca/bcl/bcla
		taken := c.branchTaken(instr)
		if !taken {
			return 3
		}
		here := c.regs.PC - 4
		target := uint32(int32(here) + fieldBD(instr))
		if fieldAA(instr) {
			target = uint32(fieldBD(instr))
		}
		if fieldLK(instr) {
			c.regs.LR = c.regs.PC
		}
		c.regs.PC = target
		return 4
	})

	setExt19(16, func(c *CPU, instr uint32) int { // bclr
		taken := c.branchTaken(instr)
		if !taken {
			return 3
		}
		target := c.regs.LR &^ 0x03
		if fieldLK(instr) {
			c.regs.LR = c.regs.PC
		}
		c.regs.PC = target
		return 4
	})

	setExt19(528, func(c *CPU, instr uint32) int { // bcctr
		taken := c.branchTaken(instr)
		if !taken {
			return 3
		}
		target := c.regs.CTR &^ 0x03
		if fieldLK(instr) {
			c.regs.LR = c.regs.PC
		}
		c.regs.PC = target
		return 4
	})
}
