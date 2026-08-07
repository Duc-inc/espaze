package cpu

func inc(c *CPU, l location) { v := c.readLoc(l) + 1; c.writeLoc(l, v); c.regs.setNZ(v) }
func dec(c *CPU, l location) { v := c.readLoc(l) - 1; c.writeLoc(l, v); c.regs.setNZ(v) }

func init() {
	setOp(0x1A, func(c *CPU) int { inc(c, c.accumulatorLoc()); return 2 }) // INC A (65C02 addition)
	setOp(0xE6, func(c *CPU) int { inc(c, c.zeroPage()); return 5 })
	setOp(0xF6, func(c *CPU) int { inc(c, c.zeroPageX()); return 6 })
	setOp(0xEE, func(c *CPU) int { inc(c, c.absolute()); return 6 })
	setOp(0xFE, func(c *CPU) int { inc(c, c.absoluteX()); return 7 })
	setOp(0xE8, func(c *CPU) int { c.regs.X++; c.regs.setNZ(c.regs.X); return 2 })
	setOp(0xC8, func(c *CPU) int { c.regs.Y++; c.regs.setNZ(c.regs.Y); return 2 })

	setOp(0x3A, func(c *CPU) int { dec(c, c.accumulatorLoc()); return 2 }) // DEC A (65C02 addition)
	setOp(0xC6, func(c *CPU) int { dec(c, c.zeroPage()); return 5 })
	setOp(0xD6, func(c *CPU) int { dec(c, c.zeroPageX()); return 6 })
	setOp(0xCE, func(c *CPU) int { dec(c, c.absolute()); return 6 })
	setOp(0xDE, func(c *CPU) int { dec(c, c.absoluteX()); return 7 })
	setOp(0xCA, func(c *CPU) int { c.regs.X--; c.regs.setNZ(c.regs.X); return 2 })
	setOp(0x88, func(c *CPU) int { c.regs.Y--; c.regs.setNZ(c.regs.Y); return 2 })
}
