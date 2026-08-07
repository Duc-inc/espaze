package cpu

func init() {
	setOp(0x48, func(c *CPU) int { c.push(c.regs.A); return 3 })
	setOp(0x08, func(c *CPU) int { c.push(c.regs.P | FlagBreak | FlagUnused); return 3 })
	setOp(0x68, func(c *CPU) int { c.regs.A = c.pop(); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0x28, func(c *CPU) int { c.regs.P = c.pop()&^FlagBreak | FlagUnused; return 4 })
}
