package cpu

func init() {
	setOp(0x18, func(c *CPU) int { c.regs.setFlag(FlagCarry, false); return 2 })
	setOp(0x38, func(c *CPU) int { c.regs.setFlag(FlagCarry, true); return 2 })
	setOp(0x58, func(c *CPU) int { c.regs.setFlag(FlagIRQD, false); return 2 })
	setOp(0x78, func(c *CPU) int { c.regs.setFlag(FlagIRQD, true); return 2 })
	setOp(0xB8, func(c *CPU) int { c.regs.setFlag(FlagOverflow, false); return 2 })
	setOp(0xD8, func(c *CPU) int { c.regs.setFlag(FlagDecimal, false); return 2 })
	setOp(0xF8, func(c *CPU) int { c.regs.setFlag(FlagDecimal, true); return 2 })

	setOp(0xC2, func(c *CPU) int { c.regs.P &^= c.fetch8(); return 3 })
	setOp(0xE2, func(c *CPU) int { c.regs.P |= c.fetch8(); return 3 })

	setOp(0xFB, func(c *CPU) int {
		oldCarry := c.regs.getFlag(FlagCarry)
		c.regs.setFlag(FlagCarry, c.regs.E)
		c.regs.E = oldCarry
		if c.regs.E {
			c.regs.setFlag(FlagIndex8, true)
			c.regs.setFlag(FlagAccum8, true)
			c.regs.X &= 0x00FF
			c.regs.Y &= 0x00FF
			c.regs.S = c.regs.S&0x00FF | 0x0100
		}
		return 2
	})

	setOp(0xEA, func(c *CPU) int { return 2 })
	setOp(0xDB, func(c *CPU) int { c.stopped = true; return 3 })
	setOp(0xCB, func(c *CPU) int { c.stopped = true; return 3 }) // WAI: simplified to behave like STP
}
