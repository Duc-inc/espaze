package cpu

func opAND(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.regs.A &= c.bus.Read(addr)
	c.regs.setZN(c.regs.A)
	return 0
}

func opORA(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.regs.A |= c.bus.Read(addr)
	c.regs.setZN(c.regs.A)
	return 0
}

func opEOR(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.regs.A ^= c.bus.Read(addr)
	c.regs.setZN(c.regs.A)
	return 0
}

// opBIT tests A & M without storing the result: Zero comes from that
// masked value, but Negative and Overflow are copied straight from bits
// 7 and 6 of M itself - the one instruction where N/V don't describe A.
func opBIT(c *CPU, _ addrMode, addr uint16, _ bool) int {
	value := c.bus.Read(addr)
	c.regs.setFlag(FlagZero, c.regs.A&value == 0)
	c.regs.setFlag(FlagNegative, value&0x80 != 0)
	c.regs.setFlag(FlagOverflow, value&0x40 != 0)
	return 0
}
