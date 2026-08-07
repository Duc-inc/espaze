package cpu

func branch(c *CPU, cond bool) int {
	target := c.relativeTarget()
	if !cond {
		return 2
	}
	c.regs.PC = target
	return 3
}

func init() {
	setOp(0x90, func(c *CPU) int { return branch(c, !c.regs.getFlag(FlagCarry)) })
	setOp(0xB0, func(c *CPU) int { return branch(c, c.regs.getFlag(FlagCarry)) })
	setOp(0xF0, func(c *CPU) int { return branch(c, c.regs.getFlag(FlagZero)) })
	setOp(0x30, func(c *CPU) int { return branch(c, c.regs.getFlag(FlagNegative)) })
	setOp(0xD0, func(c *CPU) int { return branch(c, !c.regs.getFlag(FlagZero)) })
	setOp(0x10, func(c *CPU) int { return branch(c, !c.regs.getFlag(FlagNegative)) })
	setOp(0x50, func(c *CPU) int { return branch(c, !c.regs.getFlag(FlagOverflow)) })
	setOp(0x70, func(c *CPU) int { return branch(c, c.regs.getFlag(FlagOverflow)) })
	setOp(0x80, func(c *CPU) int { return branch(c, true) }) // BRA: 65C02 addition, unconditional
}
