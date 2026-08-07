package cpu

func (c *CPU) incLoc(l location) {
	if c.regs.accum8() {
		v := c.readLoc8(l) + 1
		c.writeLoc8(l, v)
		c.regs.setNZ8(v)
	} else {
		v := c.readLoc16(l) + 1
		c.writeLoc16(l, v)
		c.regs.setNZ16(v)
	}
}

func (c *CPU) decLoc(l location) {
	if c.regs.accum8() {
		v := c.readLoc8(l) - 1
		c.writeLoc8(l, v)
		c.regs.setNZ8(v)
	} else {
		v := c.readLoc16(l) - 1
		c.writeLoc16(l, v)
		c.regs.setNZ16(v)
	}
}

func init() {
	setOp(0x1A, func(c *CPU) int { c.incLoc(c.accumulatorLoc()); return 2 })
	setOp(0xE6, func(c *CPU) int { c.incLoc(c.directPage()); return 5 })
	setOp(0xF6, func(c *CPU) int { c.incLoc(c.directPageX()); return 6 })
	setOp(0xEE, func(c *CPU) int { c.incLoc(c.absolute()); return 6 })
	setOp(0xFE, func(c *CPU) int { c.incLoc(c.absoluteX()); return 7 })

	setOp(0x3A, func(c *CPU) int { c.decLoc(c.accumulatorLoc()); return 2 })
	setOp(0xC6, func(c *CPU) int { c.decLoc(c.directPage()); return 5 })
	setOp(0xD6, func(c *CPU) int { c.decLoc(c.directPageX()); return 6 })
	setOp(0xCE, func(c *CPU) int { c.decLoc(c.absolute()); return 6 })
	setOp(0xDE, func(c *CPU) int { c.decLoc(c.absoluteX()); return 7 })

	setOp(0xE8, func(c *CPU) int { c.incIndex(&c.regs.X); return 2 })
	setOp(0xC8, func(c *CPU) int { c.incIndex(&c.regs.Y); return 2 })
	setOp(0xCA, func(c *CPU) int { c.decIndex(&c.regs.X); return 2 })
	setOp(0x88, func(c *CPU) int { c.decIndex(&c.regs.Y); return 2 })
}

func (c *CPU) incIndex(reg *uint16) {
	*reg++
	if c.regs.index8() {
		c.regs.setNZ8(byte(*reg))
	} else {
		c.regs.setNZ16(*reg)
	}
}

func (c *CPU) decIndex(reg *uint16) {
	*reg--
	if c.regs.index8() {
		c.regs.setNZ8(byte(*reg))
	} else {
		c.regs.setNZ16(*reg)
	}
}
