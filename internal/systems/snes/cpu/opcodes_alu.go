package cpu

func (c *CPU) aluAdc(l location) {
	if c.regs.accum8() {
		c.regs.A = c.regs.A&0xFF00 | uint16(c.adc8(byte(c.regs.A), c.readLoc8(l)))
	} else {
		c.regs.A = c.adc16(c.regs.A, c.readLoc16(l))
	}
}

func (c *CPU) aluSbc(l location) {
	if c.regs.accum8() {
		c.regs.A = c.regs.A&0xFF00 | uint16(c.sbc8(byte(c.regs.A), c.readLoc8(l)))
	} else {
		c.regs.A = c.sbc16(c.regs.A, c.readLoc16(l))
	}
}

func (c *CPU) aluCmp(l location) {
	if c.regs.accum8() {
		c.cmp8(byte(c.regs.A), c.readLoc8(l))
	} else {
		c.cmp16(c.regs.A, c.readLoc16(l))
	}
}

func (c *CPU) aluAnd(l location) {
	if c.regs.accum8() {
		v := byte(c.regs.A) & c.readLoc8(l)
		c.regs.A = c.regs.A&0xFF00 | uint16(v)
		c.regs.setNZ8(v)
	} else {
		v := c.regs.A & c.readLoc16(l)
		c.regs.A = v
		c.regs.setNZ16(v)
	}
}

func (c *CPU) aluOra(l location) {
	if c.regs.accum8() {
		v := byte(c.regs.A) | c.readLoc8(l)
		c.regs.A = c.regs.A&0xFF00 | uint16(v)
		c.regs.setNZ8(v)
	} else {
		v := c.regs.A | c.readLoc16(l)
		c.regs.A = v
		c.regs.setNZ16(v)
	}
}

func (c *CPU) aluEor(l location) {
	if c.regs.accum8() {
		v := byte(c.regs.A) ^ c.readLoc8(l)
		c.regs.A = c.regs.A&0xFF00 | uint16(v)
		c.regs.setNZ8(v)
	} else {
		v := c.regs.A ^ c.readLoc16(l)
		c.regs.A = v
		c.regs.setNZ16(v)
	}
}

func (c *CPU) aluBit(l location) {
	if c.regs.accum8() {
		v := c.readLoc8(l)
		c.regs.setFlag(FlagZero, byte(c.regs.A)&v == 0)
		c.regs.setFlag(FlagOverflow, v&0x40 != 0)
		c.regs.setFlag(FlagNegative, v&0x80 != 0)
	} else {
		v := c.readLoc16(l)
		c.regs.setFlag(FlagZero, c.regs.A&v == 0)
		c.regs.setFlag(FlagOverflow, v&0x4000 != 0)
		c.regs.setFlag(FlagNegative, v&0x8000 != 0)
	}
}

func init() {
	setOp(0x69, func(c *CPU) int { c.aluAdc(c.immediateLoc(c.immediateA())); return 2 })
	setOp(0x65, func(c *CPU) int { c.aluAdc(c.directPage()); return 3 })
	setOp(0x75, func(c *CPU) int { c.aluAdc(c.directPageX()); return 4 })
	setOp(0x6D, func(c *CPU) int { c.aluAdc(c.absolute()); return 4 })
	setOp(0x7D, func(c *CPU) int { c.aluAdc(c.absoluteX()); return 4 })
	setOp(0x79, func(c *CPU) int { c.aluAdc(c.absoluteY()); return 4 })
	setOp(0x6F, func(c *CPU) int { c.aluAdc(c.absoluteLong()); return 5 })
	setOp(0x71, func(c *CPU) int { c.aluAdc(c.directPageIndirectY()); return 5 })

	setOp(0xE9, func(c *CPU) int { c.aluSbc(c.immediateLoc(c.immediateA())); return 2 })
	setOp(0xE5, func(c *CPU) int { c.aluSbc(c.directPage()); return 3 })
	setOp(0xF5, func(c *CPU) int { c.aluSbc(c.directPageX()); return 4 })
	setOp(0xED, func(c *CPU) int { c.aluSbc(c.absolute()); return 4 })
	setOp(0xFD, func(c *CPU) int { c.aluSbc(c.absoluteX()); return 4 })
	setOp(0xF9, func(c *CPU) int { c.aluSbc(c.absoluteY()); return 4 })
	setOp(0xEF, func(c *CPU) int { c.aluSbc(c.absoluteLong()); return 5 })
	setOp(0xF1, func(c *CPU) int { c.aluSbc(c.directPageIndirectY()); return 5 })

	setOp(0xC9, func(c *CPU) int { c.aluCmp(c.immediateLoc(c.immediateA())); return 2 })
	setOp(0xC5, func(c *CPU) int { c.aluCmp(c.directPage()); return 3 })
	setOp(0xD5, func(c *CPU) int { c.aluCmp(c.directPageX()); return 4 })
	setOp(0xCD, func(c *CPU) int { c.aluCmp(c.absolute()); return 4 })
	setOp(0xDD, func(c *CPU) int { c.aluCmp(c.absoluteX()); return 4 })
	setOp(0xD9, func(c *CPU) int { c.aluCmp(c.absoluteY()); return 4 })

	setOp(0xE0, func(c *CPU) int {
		v := c.immediateIndex()
		if c.regs.index8() {
			c.cmp8(byte(c.regs.X), byte(v))
		} else {
			c.cmp16(c.regs.X, v)
		}
		return 2
	})
	setOp(0xC0, func(c *CPU) int {
		v := c.immediateIndex()
		if c.regs.index8() {
			c.cmp8(byte(c.regs.Y), byte(v))
		} else {
			c.cmp16(c.regs.Y, v)
		}
		return 2
	})

	setOp(0x29, func(c *CPU) int { c.aluAnd(c.immediateLoc(c.immediateA())); return 2 })
	setOp(0x25, func(c *CPU) int { c.aluAnd(c.directPage()); return 3 })
	setOp(0x2D, func(c *CPU) int { c.aluAnd(c.absolute()); return 4 })

	setOp(0x09, func(c *CPU) int { c.aluOra(c.immediateLoc(c.immediateA())); return 2 })
	setOp(0x05, func(c *CPU) int { c.aluOra(c.directPage()); return 3 })
	setOp(0x0D, func(c *CPU) int { c.aluOra(c.absolute()); return 4 })

	setOp(0x49, func(c *CPU) int { c.aluEor(c.immediateLoc(c.immediateA())); return 2 })
	setOp(0x45, func(c *CPU) int { c.aluEor(c.directPage()); return 3 })
	setOp(0x4D, func(c *CPU) int { c.aluEor(c.absolute()); return 4 })

	setOp(0x89, func(c *CPU) int { c.aluBit(c.immediateLoc(c.immediateA())); return 2 })
	setOp(0x24, func(c *CPU) int { c.aluBit(c.directPage()); return 3 })
	setOp(0x2C, func(c *CPU) int { c.aluBit(c.absolute()); return 4 })
}
