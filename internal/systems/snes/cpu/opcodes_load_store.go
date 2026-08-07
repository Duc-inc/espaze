package cpu

func (c *CPU) loadA(l location) {
	if c.regs.accum8() {
		v := c.readLoc8(l)
		c.regs.A = c.regs.A&0xFF00 | uint16(v)
		c.regs.setNZ8(v)
	} else {
		v := c.readLoc16(l)
		c.regs.A = v
		c.regs.setNZ16(v)
	}
}

func (c *CPU) storeA(l location) {
	if c.regs.accum8() {
		c.writeLoc8(l, byte(c.regs.A))
	} else {
		c.writeLoc16(l, c.regs.A)
	}
}

func (c *CPU) loadReg16(l location, dst *uint16) {
	if c.regs.index8() {
		v := c.readLoc8(l)
		*dst = uint16(v)
		c.regs.setNZ8(v)
	} else {
		v := c.readLoc16(l)
		*dst = v
		c.regs.setNZ16(v)
	}
}

func (c *CPU) storeReg16(l location, v uint16) {
	if c.regs.index8() {
		c.writeLoc8(l, byte(v))
	} else {
		c.writeLoc16(l, v)
	}
}

// immediateA/immediateIndex fetch an operand sized to A's or X/Y's
// current 8/16-bit width, matching real hardware's own variable-length
// immediate operand.
func (c *CPU) immediateA() uint16 {
	if c.regs.accum8() {
		return uint16(c.fetch8())
	}
	return c.fetch16()
}

func (c *CPU) immediateIndex() uint16 {
	if c.regs.index8() {
		return uint16(c.fetch8())
	}
	return c.fetch16()
}

func init() {
	setOp(0xA9, func(c *CPU) int { v := c.immediateA(); c.applyLoadA(v); return 2 })
	setOp(0xA5, func(c *CPU) int { c.loadA(c.directPage()); return 3 })
	setOp(0xB5, func(c *CPU) int { c.loadA(c.directPageX()); return 4 })
	setOp(0xAD, func(c *CPU) int { c.loadA(c.absolute()); return 4 })
	setOp(0xBD, func(c *CPU) int { c.loadA(c.absoluteX()); return 4 })
	setOp(0xB9, func(c *CPU) int { c.loadA(c.absoluteY()); return 4 })
	setOp(0xAF, func(c *CPU) int { c.loadA(c.absoluteLong()); return 5 })
	setOp(0xBF, func(c *CPU) int { c.loadA(c.absoluteLongX()); return 5 })
	setOp(0xB2, func(c *CPU) int { c.loadA(c.directPageIndirect()); return 5 })
	setOp(0xB1, func(c *CPU) int { c.loadA(c.directPageIndirectY()); return 5 })
	setOp(0xA7, func(c *CPU) int { c.loadA(c.directPageIndirectLong()); return 6 })
	setOp(0xB7, func(c *CPU) int { c.loadA(c.directPageIndirectLongY()); return 6 })

	setOp(0x85, func(c *CPU) int { c.storeA(c.directPage()); return 3 })
	setOp(0x95, func(c *CPU) int { c.storeA(c.directPageX()); return 4 })
	setOp(0x8D, func(c *CPU) int { c.storeA(c.absolute()); return 4 })
	setOp(0x9D, func(c *CPU) int { c.storeA(c.absoluteX()); return 4 })
	setOp(0x99, func(c *CPU) int { c.storeA(c.absoluteY()); return 4 })
	setOp(0x8F, func(c *CPU) int { c.storeA(c.absoluteLong()); return 5 })
	setOp(0x9F, func(c *CPU) int { c.storeA(c.absoluteLongX()); return 5 })
	setOp(0x92, func(c *CPU) int { c.storeA(c.directPageIndirect()); return 5 })
	setOp(0x91, func(c *CPU) int { c.storeA(c.directPageIndirectY()); return 5 })
	setOp(0x87, func(c *CPU) int { c.storeA(c.directPageIndirectLong()); return 6 })
	setOp(0x97, func(c *CPU) int { c.storeA(c.directPageIndirectLongY()); return 6 })

	setOp(0xA2, func(c *CPU) int { v := c.immediateIndex(); c.applyLoadX(v); return 2 })
	setOp(0xA6, func(c *CPU) int { c.loadReg16(c.directPage(), &c.regs.X); return 3 })
	setOp(0xB6, func(c *CPU) int { c.loadReg16(c.directPageY(), &c.regs.X); return 4 })
	setOp(0xAE, func(c *CPU) int { c.loadReg16(c.absolute(), &c.regs.X); return 4 })
	setOp(0xBE, func(c *CPU) int { c.loadReg16(c.absoluteY(), &c.regs.X); return 4 })

	setOp(0x86, func(c *CPU) int { c.storeReg16(c.directPage(), c.regs.X); return 3 })
	setOp(0x96, func(c *CPU) int { c.storeReg16(c.directPageY(), c.regs.X); return 4 })
	setOp(0x8E, func(c *CPU) int { c.storeReg16(c.absolute(), c.regs.X); return 4 })

	setOp(0xA0, func(c *CPU) int { v := c.immediateIndex(); c.applyLoadY(v); return 2 })
	setOp(0xA4, func(c *CPU) int { c.loadReg16(c.directPage(), &c.regs.Y); return 3 })
	setOp(0xB4, func(c *CPU) int { c.loadReg16(c.directPageX(), &c.regs.Y); return 4 })
	setOp(0xAC, func(c *CPU) int { c.loadReg16(c.absolute(), &c.regs.Y); return 4 })
	setOp(0xBC, func(c *CPU) int { c.loadReg16(c.absoluteX(), &c.regs.Y); return 4 })

	setOp(0x84, func(c *CPU) int { c.storeReg16(c.directPage(), c.regs.Y); return 3 })
	setOp(0x94, func(c *CPU) int { c.storeReg16(c.directPageX(), c.regs.Y); return 4 })
	setOp(0x8C, func(c *CPU) int { c.storeReg16(c.absolute(), c.regs.Y); return 4 })

	setOp(0x64, func(c *CPU) int { c.storeZero(c.directPage()); return 3 })
	setOp(0x74, func(c *CPU) int { c.storeZero(c.directPageX()); return 4 })
	setOp(0x9C, func(c *CPU) int { c.storeZero(c.absolute()); return 4 })
	setOp(0x9E, func(c *CPU) int { c.storeZero(c.absoluteX()); return 4 })
}

func (c *CPU) applyLoadA(v uint16) {
	if c.regs.accum8() {
		c.regs.A = c.regs.A&0xFF00 | v&0xFF
		c.regs.setNZ8(byte(v))
	} else {
		c.regs.A = v
		c.regs.setNZ16(v)
	}
}

func (c *CPU) applyLoadX(v uint16) {
	c.regs.X = v
	if c.regs.index8() {
		c.regs.setNZ8(byte(v))
	} else {
		c.regs.setNZ16(v)
	}
}

func (c *CPU) applyLoadY(v uint16) {
	c.regs.Y = v
	if c.regs.index8() {
		c.regs.setNZ8(byte(v))
	} else {
		c.regs.setNZ16(v)
	}
}

func (c *CPU) storeZero(l location) {
	if c.regs.accum8() {
		c.writeLoc8(l, 0)
	} else {
		c.writeLoc16(l, 0)
	}
}
