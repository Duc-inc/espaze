package cpu

func opINC(c *CPU, _ addrMode, addr uint16, _ bool) int {
	value := c.bus.Read(addr) + 1
	c.bus.Write(addr, value)
	c.regs.setZN(value)
	return 0
}

func opDEC(c *CPU, _ addrMode, addr uint16, _ bool) int {
	value := c.bus.Read(addr) - 1
	c.bus.Write(addr, value)
	c.regs.setZN(value)
	return 0
}

func opINX(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.X++
	c.regs.setZN(c.regs.X)
	return 0
}

func opINY(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.Y++
	c.regs.setZN(c.regs.Y)
	return 0
}

func opDEX(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.X--
	c.regs.setZN(c.regs.X)
	return 0
}

func opDEY(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.Y--
	c.regs.setZN(c.regs.Y)
	return 0
}
