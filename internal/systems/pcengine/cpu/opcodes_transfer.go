package cpu

func init() {
	setOp(0xAA, func(c *CPU) int { c.regs.X = c.regs.A; c.regs.setNZ(c.regs.X); return 2 })
	setOp(0xA8, func(c *CPU) int { c.regs.Y = c.regs.A; c.regs.setNZ(c.regs.Y); return 2 })
	setOp(0xBA, func(c *CPU) int { c.regs.X = c.regs.S; c.regs.setNZ(c.regs.X); return 2 })
	setOp(0x8A, func(c *CPU) int { c.regs.A = c.regs.X; c.regs.setNZ(c.regs.A); return 2 })
	setOp(0x9A, func(c *CPU) int { c.regs.S = c.regs.X; return 2 })
	setOp(0x98, func(c *CPU) int { c.regs.A = c.regs.Y; c.regs.setNZ(c.regs.A); return 2 })

	// PHX/PLX/PHY/PLY: 65C02 additions.
	setOp(0xDA, func(c *CPU) int { c.push(c.regs.X); return 3 })
	setOp(0xFA, func(c *CPU) int { c.regs.X = c.pop(); c.regs.setNZ(c.regs.X); return 4 })
	setOp(0x5A, func(c *CPU) int { c.push(c.regs.Y); return 3 })
	setOp(0x7A, func(c *CPU) int { c.regs.Y = c.pop(); c.regs.setNZ(c.regs.Y); return 4 })
}
