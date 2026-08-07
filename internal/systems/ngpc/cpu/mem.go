package cpu

func (c *CPU) fetch8() byte {
	v := c.bus.Read8(c.regs.PC)
	c.regs.PC++
	return v
}

func (c *CPU) fetch16() uint16 {
	v := c.bus.Read16(c.regs.PC)
	c.regs.PC += 2
	return v
}

func (c *CPU) fetch32() uint32 {
	lo := uint32(c.fetch16())
	hi := uint32(c.fetch16())
	return lo | hi<<16
}

func (c *CPU) push16(v uint16) {
	c.regs.XSP -= 2
	c.bus.Write16(c.regs.XSP, v)
}

func (c *CPU) pop16() uint16 {
	v := c.bus.Read16(c.regs.XSP)
	c.regs.XSP += 2
	return v
}

func (c *CPU) push32(v uint32) {
	c.regs.XSP -= 4
	c.bus.Write16(c.regs.XSP, uint16(v))
	c.bus.Write16(c.regs.XSP+2, uint16(v>>16))
}

func (c *CPU) pop32() uint32 {
	lo := uint32(c.bus.Read16(c.regs.XSP))
	hi := uint32(c.bus.Read16(c.regs.XSP + 2))
	c.regs.XSP += 4
	return lo | hi<<16
}
