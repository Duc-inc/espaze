package cpu

func init() {
	setOp(0x90, func(c *CPU) int { c.regs.PC = uint32(c.fetch16()); return 5 })
	for cc := byte(0); cc < 7; cc++ {
		cond := cc
		setOp(0x91+cond, func(c *CPU) int {
			target := uint32(c.fetch16())
			if c.checkCondition(cond) {
				c.regs.PC = target
			}
			return 4
		})
	}

	setOp(0xA0, func(c *CPU) int {
		offset := int32(int8(c.fetch8()))
		c.regs.PC = uint32(int32(c.regs.PC) + offset)
		return 4
	})
	for cc := byte(0); cc < 7; cc++ {
		cond := cc
		setOp(0xA1+cond, func(c *CPU) int {
			offset := int32(int8(c.fetch8()))
			if c.checkCondition(cond) {
				c.regs.PC = uint32(int32(c.regs.PC) + offset)
			}
			return 3
		})
	}

	setOp(0xB0, func(c *CPU) int {
		target := uint32(c.fetch16())
		c.push32(c.regs.PC)
		c.regs.PC = target
		return 6
	})
	for cc := byte(0); cc < 7; cc++ {
		cond := cc
		setOp(0xB1+cond, func(c *CPU) int {
			target := uint32(c.fetch16())
			if c.checkCondition(cond) {
				c.push32(c.regs.PC)
				c.regs.PC = target
			}
			return 5
		})
	}

	setOp(0xB8, func(c *CPU) int { c.regs.PC = c.pop32(); return 5 })
	setOp(0xB9, func(c *CPU) int {
		if c.checkCondition(0) {
			c.regs.PC = c.pop32()
		}
		return 4
	})
	setOp(0xBF, func(c *CPU) int {
		c.regs.SR = c.pop16()
		c.regs.PC = c.pop32()
		return 7
	})

	for p := byte(0); p < 4; p++ {
		pair := p
		setOp(0xC0+pair, func(c *CPU) int { c.push16(c.regs.reg16(pair)); return 3 })
		setOp(0xC4+pair, func(c *CPU) int {
			c.regs.setReg16(pair, c.pop16())
			return 3
		})
	}
}
