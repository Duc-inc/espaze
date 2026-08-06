package cpu

func opTAX(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.X = c.regs.A
	c.regs.setZN(c.regs.X)
	return 0
}

func opTAY(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.Y = c.regs.A
	c.regs.setZN(c.regs.Y)
	return 0
}

func opTXA(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.A = c.regs.X
	c.regs.setZN(c.regs.A)
	return 0
}

func opTYA(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.A = c.regs.Y
	c.regs.setZN(c.regs.A)
	return 0
}

func opTSX(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.X = c.regs.SP
	c.regs.setZN(c.regs.X)
	return 0
}

func opTXS(c *CPU, _ addrMode, _ uint16, _ bool) int {
	// Unlike the other transfers, TXS doesn't touch the flags - SP isn't
	// a "value" register in the same sense.
	c.regs.SP = c.regs.X
	return 0
}
