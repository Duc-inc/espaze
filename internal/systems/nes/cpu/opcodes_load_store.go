package cpu

func opLDA(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.regs.A = c.bus.Read(addr)
	c.regs.setZN(c.regs.A)
	return 0
}

func opLDX(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.regs.X = c.bus.Read(addr)
	c.regs.setZN(c.regs.X)
	return 0
}

func opLDY(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.regs.Y = c.bus.Read(addr)
	c.regs.setZN(c.regs.Y)
	return 0
}

func opSTA(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.bus.Write(addr, c.regs.A)
	return 0
}

func opSTX(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.bus.Write(addr, c.regs.X)
	return 0
}

func opSTY(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.bus.Write(addr, c.regs.Y)
	return 0
}
