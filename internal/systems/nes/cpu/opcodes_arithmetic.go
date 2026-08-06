package cpu

func opADC(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.addWithCarry(c.bus.Read(addr))
	return 0
}

// opSBC is ADC with the operand's bits flipped: A - M - (1-carry) is
// exactly A + (^M) + carry in two's complement, so it reuses the same
// carry/overflow math instead of a separate implementation.
func opSBC(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.addWithCarry(^c.bus.Read(addr))
	return 0
}

func (c *CPU) addWithCarry(m byte) {
	a := c.regs.A
	carryIn := uint16(0)
	if c.regs.getFlag(FlagCarry) {
		carryIn = 1
	}
	sum := uint16(a) + uint16(m) + carryIn
	result := byte(sum)

	c.regs.setFlag(FlagCarry, sum > 0xFF)
	c.regs.setFlag(FlagOverflow, (a^m)&0x80 == 0 && (a^result)&0x80 != 0)
	c.regs.setZN(result)
	c.regs.A = result
}

func opCMP(c *CPU, _ addrMode, addr uint16, _ bool) int {
	compare(c, c.regs.A, c.bus.Read(addr))
	return 0
}

func opCPX(c *CPU, _ addrMode, addr uint16, _ bool) int {
	compare(c, c.regs.X, c.bus.Read(addr))
	return 0
}

func opCPY(c *CPU, _ addrMode, addr uint16, _ bool) int {
	compare(c, c.regs.Y, c.bus.Read(addr))
	return 0
}

func compare(c *CPU, reg, value byte) {
	result := reg - value
	c.regs.setFlag(FlagCarry, reg >= value)
	c.regs.setZN(result)
}
