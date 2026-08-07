package cpu

func init() {
	setOp(0x18, func(c *CPU) int { c.regs.setFlag(FlagCarry, false); return 2 })
	setOp(0xD8, func(c *CPU) int { c.regs.setFlag(FlagDecimal, false); return 2 })
	setOp(0x58, func(c *CPU) int { c.regs.setFlag(FlagInterrupt, false); return 2 })
	setOp(0xB8, func(c *CPU) int { c.regs.setFlag(FlagOverflow, false); return 2 })
	setOp(0x38, func(c *CPU) int { c.regs.setFlag(FlagCarry, true); return 2 })
	setOp(0xF8, func(c *CPU) int { c.regs.setFlag(FlagDecimal, true); return 2 })
	setOp(0x78, func(c *CPU) int { c.regs.setFlag(FlagInterrupt, true); return 2 })
}
