package cpu

func (c *CPU) pushA() {
	if c.regs.accum8() {
		c.push8(byte(c.regs.A))
	} else {
		c.push16(c.regs.A)
	}
}

func (c *CPU) pullA() {
	if c.regs.accum8() {
		v := c.pop8()
		c.regs.A = c.regs.A&0xFF00 | uint16(v)
		c.regs.setNZ8(v)
	} else {
		v := c.pop16()
		c.regs.A = v
		c.regs.setNZ16(v)
	}
}

func (c *CPU) pushIndex(v uint16) {
	if c.regs.index8() {
		c.push8(byte(v))
	} else {
		c.push16(v)
	}
}

func (c *CPU) pullIndex(reg *uint16) {
	if c.regs.index8() {
		v := c.pop8()
		*reg = uint16(v)
		c.regs.setNZ8(v)
	} else {
		v := c.pop16()
		*reg = v
		c.regs.setNZ16(v)
	}
}

func init() {
	setOp(0x48, func(c *CPU) int { c.pushA(); return 3 })
	setOp(0x68, func(c *CPU) int { c.pullA(); return 4 })
	setOp(0xDA, func(c *CPU) int { c.pushIndex(c.regs.X); return 3 })
	setOp(0xFA, func(c *CPU) int { c.pullIndex(&c.regs.X); return 4 })
	setOp(0x5A, func(c *CPU) int { c.pushIndex(c.regs.Y); return 3 })
	setOp(0x7A, func(c *CPU) int { c.pullIndex(&c.regs.Y); return 4 })
	setOp(0x08, func(c *CPU) int { c.push8(c.regs.P); return 3 })
	setOp(0x28, func(c *CPU) int { c.regs.P = c.pop8(); return 4 })
	setOp(0x8B, func(c *CPU) int { c.push8(c.regs.DBR); return 3 })
	setOp(0xAB, func(c *CPU) int { c.regs.DBR = c.pop8(); c.regs.setNZ8(c.regs.DBR); return 4 })
	setOp(0x0B, func(c *CPU) int { c.push16(c.regs.D); return 4 })
	setOp(0x2B, func(c *CPU) int { c.regs.D = c.pop16(); c.regs.setNZ16(c.regs.D); return 5 })
	setOp(0x4B, func(c *CPU) int { c.push8(c.regs.PBR); return 3 })

	setOp(0xF4, func(c *CPU) int { c.push16(c.fetch16()); return 5 }) // PEA
	setOp(0xD4, func(c *CPU) int {
		l := c.directPage()
		c.push16(c.readLoc16(l))
		return 6
	}) // PEI
	setOp(0x62, func(c *CPU) int { c.push16(c.relativeTarget16()); return 6 }) // PER
}
