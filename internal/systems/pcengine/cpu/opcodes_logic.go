package cpu

func bitTest(c *CPU, v byte) {
	c.regs.setFlag(FlagZero, c.regs.A&v == 0)
	c.regs.setFlag(FlagOverflow, v&0x40 != 0)
	c.regs.setFlag(FlagNegative, v&0x80 != 0)
}

func init() {
	setOp(0x29, func(c *CPU) int { c.regs.A &= c.readLoc(c.immediate()); c.regs.setNZ(c.regs.A); return 2 })
	setOp(0x25, func(c *CPU) int { c.regs.A &= c.readLoc(c.zeroPage()); c.regs.setNZ(c.regs.A); return 3 })
	setOp(0x35, func(c *CPU) int { c.regs.A &= c.readLoc(c.zeroPageX()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x2D, func(c *CPU) int { c.regs.A &= c.readLoc(c.absolute()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x3D, func(c *CPU) int { c.regs.A &= c.readLoc(c.absoluteX()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x39, func(c *CPU) int { c.regs.A &= c.readLoc(c.absoluteY()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x21, func(c *CPU) int { c.regs.A &= c.readLoc(c.indexedIndirect()); c.regs.setNZ(c.regs.A); return 6 })
	setOp(0x31, func(c *CPU) int { c.regs.A &= c.readLoc(c.indirectIndexed()); c.regs.setNZ(c.regs.A); return 5 })

	setOp(0x09, func(c *CPU) int { c.regs.A |= c.readLoc(c.immediate()); c.regs.setNZ(c.regs.A); return 2 })
	setOp(0x05, func(c *CPU) int { c.regs.A |= c.readLoc(c.zeroPage()); c.regs.setNZ(c.regs.A); return 3 })
	setOp(0x15, func(c *CPU) int { c.regs.A |= c.readLoc(c.zeroPageX()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x0D, func(c *CPU) int { c.regs.A |= c.readLoc(c.absolute()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x1D, func(c *CPU) int { c.regs.A |= c.readLoc(c.absoluteX()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x19, func(c *CPU) int { c.regs.A |= c.readLoc(c.absoluteY()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x01, func(c *CPU) int { c.regs.A |= c.readLoc(c.indexedIndirect()); c.regs.setNZ(c.regs.A); return 6 })
	setOp(0x11, func(c *CPU) int { c.regs.A |= c.readLoc(c.indirectIndexed()); c.regs.setNZ(c.regs.A); return 5 })

	setOp(0x49, func(c *CPU) int { c.regs.A ^= c.readLoc(c.immediate()); c.regs.setNZ(c.regs.A); return 2 })
	setOp(0x45, func(c *CPU) int { c.regs.A ^= c.readLoc(c.zeroPage()); c.regs.setNZ(c.regs.A); return 3 })
	setOp(0x55, func(c *CPU) int { c.regs.A ^= c.readLoc(c.zeroPageX()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x4D, func(c *CPU) int { c.regs.A ^= c.readLoc(c.absolute()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x5D, func(c *CPU) int { c.regs.A ^= c.readLoc(c.absoluteX()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x59, func(c *CPU) int { c.regs.A ^= c.readLoc(c.absoluteY()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x41, func(c *CPU) int { c.regs.A ^= c.readLoc(c.indexedIndirect()); c.regs.setNZ(c.regs.A); return 6 })
	setOp(0x51, func(c *CPU) int { c.regs.A ^= c.readLoc(c.indirectIndexed()); c.regs.setNZ(c.regs.A); return 5 })

	setOp(0x24, func(c *CPU) int { bitTest(c, c.readLoc(c.zeroPage())); return 3 })
	setOp(0x2C, func(c *CPU) int { bitTest(c, c.readLoc(c.absolute())); return 4 })
	setOp(0x89, func(c *CPU) int { // BIT #imm: only sets Zero, not N/V (65C02 addition)
		v := c.readLoc(c.immediate())
		c.regs.setFlag(FlagZero, c.regs.A&v == 0)
		return 2
	})

	// TRB/TSB: 65C02 additions - test-and-reset/set bits in memory
	// using A as the mask, without loading A.
	setOp(0x14, func(c *CPU) int {
		l := c.zeroPage()
		v := c.readLoc(l)
		c.regs.setFlag(FlagZero, c.regs.A&v == 0)
		c.writeLoc(l, v&^c.regs.A)
		return 5
	})
	setOp(0x1C, func(c *CPU) int {
		l := c.absolute()
		v := c.readLoc(l)
		c.regs.setFlag(FlagZero, c.regs.A&v == 0)
		c.writeLoc(l, v&^c.regs.A)
		return 6
	})
	setOp(0x04, func(c *CPU) int {
		l := c.zeroPage()
		v := c.readLoc(l)
		c.regs.setFlag(FlagZero, c.regs.A&v == 0)
		c.writeLoc(l, v|c.regs.A)
		return 5
	})
	setOp(0x0C, func(c *CPU) int {
		l := c.absolute()
		v := c.readLoc(l)
		c.regs.setFlag(FlagZero, c.regs.A&v == 0)
		c.writeLoc(l, v|c.regs.A)
		return 6
	})
}
