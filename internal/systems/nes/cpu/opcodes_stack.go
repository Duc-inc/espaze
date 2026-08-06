package cpu

func opPHA(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.push(c.regs.A)
	return 0
}

func opPHP(c *CPU, _ addrMode, _ uint16, _ bool) int {
	// The byte pushed always has Break and Unused set, even though
	// neither bit exists as real flip-flops in the status register.
	c.push(c.regs.P | FlagBreak | FlagUnused)
	return 0
}

func opPLA(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.A = c.pop()
	c.regs.setZN(c.regs.A)
	return 0
}

func opPLP(c *CPU, _ addrMode, _ uint16, _ bool) int {
	// Break has no physical storage and Unused always reads as 1;
	// neither should actually change based on what was pushed.
	c.regs.P = (c.pop() &^ FlagBreak) | FlagUnused
	return 0
}
