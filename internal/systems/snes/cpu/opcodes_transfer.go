package cpu

func init() {
	setOp(0xAA, func(c *CPU) int { c.transferToIndex(c.regs.A, &c.regs.X); return 2 })
	setOp(0xA8, func(c *CPU) int { c.transferToIndex(c.regs.A, &c.regs.Y); return 2 })
	setOp(0x8A, func(c *CPU) int { c.transferToA(c.regs.X); return 2 })
	setOp(0x98, func(c *CPU) int { c.transferToA(c.regs.Y); return 2 })
	setOp(0xBA, func(c *CPU) int { c.transferToIndex(c.regs.S, &c.regs.X); return 2 })
	setOp(0x9A, func(c *CPU) int {
		c.regs.S = c.regs.X
		if c.regs.E {
			c.regs.S = c.regs.S&0x00FF | 0x0100
		}
		return 2
	})
	setOp(0x9B, func(c *CPU) int { c.transferToIndex(c.regs.X, &c.regs.Y); return 2 })
	setOp(0xBB, func(c *CPU) int { c.transferToIndex(c.regs.Y, &c.regs.X); return 2 })

	// TCD/TDC/TCS/TSC always move the full 16-bit accumulator/stack
	// pointer, regardless of the M flag - real hardware's own rule.
	setOp(0x5B, func(c *CPU) int { c.regs.D = c.regs.A; c.regs.setNZ16(c.regs.D); return 2 })
	setOp(0x7B, func(c *CPU) int { c.regs.A = c.regs.D; c.regs.setNZ16(c.regs.A); return 2 })
	setOp(0x1B, func(c *CPU) int {
		c.regs.S = c.regs.A
		if c.regs.E {
			c.regs.S = c.regs.S&0x00FF | 0x0100
		}
		return 2
	})
	setOp(0x3B, func(c *CPU) int { c.regs.A = c.regs.S; c.regs.setNZ16(c.regs.A); return 2 })
}

func (c *CPU) transferToIndex(src uint16, dst *uint16) {
	if c.regs.index8() {
		*dst = src & 0x00FF
		c.regs.setNZ8(byte(*dst))
	} else {
		*dst = src
		c.regs.setNZ16(*dst)
	}
}

func (c *CPU) transferToA(src uint16) {
	if c.regs.accum8() {
		c.regs.A = c.regs.A&0xFF00 | src&0x00FF
		c.regs.setNZ8(byte(c.regs.A))
	} else {
		c.regs.A = src
		c.regs.setNZ16(c.regs.A)
	}
}
