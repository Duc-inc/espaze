package cpu

func init() {
	setOp(0x4C, func(c *CPU) int { c.regs.PC = c.fetch16(); return 3 })
	setOp(0x6C, func(c *CPU) int { // JMP (abs)
		ptr := c.fetch16()
		lo := uint16(c.read(ptr))
		hi := uint16(c.read(ptr + 1))
		c.regs.PC = lo | hi<<8
		return 5
	})
	setOp(0x7C, func(c *CPU) int { // JMP (abs,X): 65C02 addition
		base := c.fetch16()
		ptr := base + uint16(c.regs.X)
		lo := uint16(c.read(ptr))
		hi := uint16(c.read(ptr + 1))
		c.regs.PC = lo | hi<<8
		return 6
	})
	setOp(0x20, func(c *CPU) int {
		target := c.fetch16()
		c.push16(c.regs.PC - 1)
		c.regs.PC = target
		return 6
	})
	setOp(0x60, func(c *CPU) int { c.regs.PC = c.pop16() + 1; return 6 })
	setOp(0x40, func(c *CPU) int {
		c.regs.P = c.pop()&^FlagBreak | FlagUnused
		c.regs.PC = c.pop16()
		return 6
	})
	setOp(0x00, func(c *CPU) int {
		c.regs.PC++
		c.serviceInterrupt(0xFFF2, true) // HuC6280 BRK vector (distinct from IRQ vectors)
		return 7
	})
	setOp(0xEA, func(c *CPU) int { return 2 }) // NOP
}
