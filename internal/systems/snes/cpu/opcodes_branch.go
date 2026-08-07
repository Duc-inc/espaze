package cpu

func init() {
	conditions := []byte{0x10, 0x30, 0x50, 0x70, 0x90, 0xB0, 0xD0, 0xF0}
	for i, opcode := range conditions {
		cc := byte(i)
		setOp(opcode, func(c *CPU) int {
			target := c.relativeTarget8()
			if c.checkCondition(cc) {
				c.regs.PC = target
				return 3
			}
			return 2
		})
	}

	setOp(0x80, func(c *CPU) int { c.regs.PC = c.relativeTarget8(); return 3 })  // BRA
	setOp(0x82, func(c *CPU) int { c.regs.PC = c.relativeTarget16(); return 4 }) // BRL
}
