package cpu

func (c *CPU) read8(addr uint32) byte        { return c.bus.Read8(addr) }
func (c *CPU) read16(addr uint32) uint16     { return c.bus.Read16(addr) }
func (c *CPU) read32(addr uint32) uint32     { return c.bus.Read32(addr) }
func (c *CPU) write8(addr uint32, v byte)    { c.bus.Write8(addr, v) }
func (c *CPU) write16(addr uint32, v uint16) { c.bus.Write16(addr, v) }
func (c *CPU) write32(addr uint32, v uint32) { c.bus.Write32(addr, v) }

func (c *CPU) fetch16() uint16 {
	v := c.read16(c.regs.R[15])
	c.regs.R[15] += 2
	return v
}

func (c *CPU) fetch32() uint32 {
	v := c.read32(c.regs.R[15])
	c.regs.R[15] += 4
	return v
}

func (c *CPU) push(v uint32) {
	c.regs.R[13] -= 4
	c.write32(c.regs.R[13], v)
}

func (c *CPU) pop() uint32 {
	v := c.read32(c.regs.R[13])
	c.regs.R[13] += 4
	return v
}
