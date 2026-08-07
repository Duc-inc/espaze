package cpu

func asl(c *CPU, l location) {
	v := c.readLoc(l)
	c.regs.setFlag(FlagCarry, v&0x80 != 0)
	v <<= 1
	c.writeLoc(l, v)
	c.regs.setNZ(v)
}

func lsr(c *CPU, l location) {
	v := c.readLoc(l)
	c.regs.setFlag(FlagCarry, v&0x01 != 0)
	v >>= 1
	c.writeLoc(l, v)
	c.regs.setNZ(v)
}

func rol(c *CPU, l location) {
	v := c.readLoc(l)
	carryIn := byte(0)
	if c.regs.getFlag(FlagCarry) {
		carryIn = 1
	}
	c.regs.setFlag(FlagCarry, v&0x80 != 0)
	v = v<<1 | carryIn
	c.writeLoc(l, v)
	c.regs.setNZ(v)
}

func ror(c *CPU, l location) {
	v := c.readLoc(l)
	carryIn := byte(0)
	if c.regs.getFlag(FlagCarry) {
		carryIn = 0x80
	}
	c.regs.setFlag(FlagCarry, v&0x01 != 0)
	v = v>>1 | carryIn
	c.writeLoc(l, v)
	c.regs.setNZ(v)
}

func init() {
	setOp(0x0A, func(c *CPU) int { asl(c, c.accumulatorLoc()); return 2 })
	setOp(0x06, func(c *CPU) int { asl(c, c.zeroPage()); return 5 })
	setOp(0x16, func(c *CPU) int { asl(c, c.zeroPageX()); return 6 })
	setOp(0x0E, func(c *CPU) int { asl(c, c.absolute()); return 6 })
	setOp(0x1E, func(c *CPU) int { asl(c, c.absoluteX()); return 7 })

	setOp(0x4A, func(c *CPU) int { lsr(c, c.accumulatorLoc()); return 2 })
	setOp(0x46, func(c *CPU) int { lsr(c, c.zeroPage()); return 5 })
	setOp(0x56, func(c *CPU) int { lsr(c, c.zeroPageX()); return 6 })
	setOp(0x4E, func(c *CPU) int { lsr(c, c.absolute()); return 6 })
	setOp(0x5E, func(c *CPU) int { lsr(c, c.absoluteX()); return 7 })

	setOp(0x2A, func(c *CPU) int { rol(c, c.accumulatorLoc()); return 2 })
	setOp(0x26, func(c *CPU) int { rol(c, c.zeroPage()); return 5 })
	setOp(0x36, func(c *CPU) int { rol(c, c.zeroPageX()); return 6 })
	setOp(0x2E, func(c *CPU) int { rol(c, c.absolute()); return 6 })
	setOp(0x3E, func(c *CPU) int { rol(c, c.absoluteX()); return 7 })

	setOp(0x6A, func(c *CPU) int { ror(c, c.accumulatorLoc()); return 2 })
	setOp(0x66, func(c *CPU) int { ror(c, c.zeroPage()); return 5 })
	setOp(0x76, func(c *CPU) int { ror(c, c.zeroPageX()); return 6 })
	setOp(0x6E, func(c *CPU) int { ror(c, c.absolute()); return 6 })
	setOp(0x7E, func(c *CPU) int { ror(c, c.absoluteX()); return 7 })
}
