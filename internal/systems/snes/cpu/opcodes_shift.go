package cpu

func (c *CPU) shift(l location, op8 func(byte) byte, op16 func(uint16) uint16) {
	if c.regs.accum8() {
		c.writeLoc8(l, op8(c.readLoc8(l)))
	} else {
		c.writeLoc16(l, op16(c.readLoc16(l)))
	}
}

func init() {
	setOp(0x0A, func(c *CPU) int { c.shift(c.accumulatorLoc(), c.asl8, c.asl16); return 2 })
	setOp(0x06, func(c *CPU) int { c.shift(c.directPage(), c.asl8, c.asl16); return 5 })
	setOp(0x16, func(c *CPU) int { c.shift(c.directPageX(), c.asl8, c.asl16); return 6 })
	setOp(0x0E, func(c *CPU) int { c.shift(c.absolute(), c.asl8, c.asl16); return 6 })
	setOp(0x1E, func(c *CPU) int { c.shift(c.absoluteX(), c.asl8, c.asl16); return 7 })

	setOp(0x4A, func(c *CPU) int { c.shift(c.accumulatorLoc(), c.lsr8, c.lsr16); return 2 })
	setOp(0x46, func(c *CPU) int { c.shift(c.directPage(), c.lsr8, c.lsr16); return 5 })
	setOp(0x56, func(c *CPU) int { c.shift(c.directPageX(), c.lsr8, c.lsr16); return 6 })
	setOp(0x4E, func(c *CPU) int { c.shift(c.absolute(), c.lsr8, c.lsr16); return 6 })
	setOp(0x5E, func(c *CPU) int { c.shift(c.absoluteX(), c.lsr8, c.lsr16); return 7 })

	setOp(0x2A, func(c *CPU) int { c.shift(c.accumulatorLoc(), c.rol8, c.rol16); return 2 })
	setOp(0x26, func(c *CPU) int { c.shift(c.directPage(), c.rol8, c.rol16); return 5 })
	setOp(0x36, func(c *CPU) int { c.shift(c.directPageX(), c.rol8, c.rol16); return 6 })
	setOp(0x2E, func(c *CPU) int { c.shift(c.absolute(), c.rol8, c.rol16); return 6 })
	setOp(0x3E, func(c *CPU) int { c.shift(c.absoluteX(), c.rol8, c.rol16); return 7 })

	setOp(0x6A, func(c *CPU) int { c.shift(c.accumulatorLoc(), c.ror8, c.ror16); return 2 })
	setOp(0x66, func(c *CPU) int { c.shift(c.directPage(), c.ror8, c.ror16); return 5 })
	setOp(0x76, func(c *CPU) int { c.shift(c.directPageX(), c.ror8, c.ror16); return 6 })
	setOp(0x6E, func(c *CPU) int { c.shift(c.absolute(), c.ror8, c.ror16); return 6 })
	setOp(0x7E, func(c *CPU) int { c.shift(c.absoluteX(), c.ror8, c.ror16); return 7 })
}
