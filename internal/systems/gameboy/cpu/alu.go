package cpu

// 8-bit ALU helpers shared by the opcode tables. Each sets flags exactly
// as the LR35902 manual specifies for that operation.

func (c *CPU) add8(a, b byte) byte {
	sum := int(a) + int(b)
	result := byte(sum)
	c.regs.SetFlag(FlagZ, result == 0)
	c.regs.SetFlag(FlagN, false)
	c.regs.SetFlag(FlagH, (a&0xF)+(b&0xF) > 0xF)
	c.regs.SetFlag(FlagC, sum > 0xFF)
	return result
}

func (c *CPU) adc8(a, b byte) byte {
	carry := c.carryIn()
	sum := int(a) + int(b) + int(carry)
	result := byte(sum)
	c.regs.SetFlag(FlagZ, result == 0)
	c.regs.SetFlag(FlagN, false)
	c.regs.SetFlag(FlagH, (a&0xF)+(b&0xF)+carry > 0xF)
	c.regs.SetFlag(FlagC, sum > 0xFF)
	return result
}

func (c *CPU) sub8(a, b byte) byte {
	result := a - b
	c.regs.SetFlag(FlagZ, result == 0)
	c.regs.SetFlag(FlagN, true)
	c.regs.SetFlag(FlagH, a&0xF < b&0xF)
	c.regs.SetFlag(FlagC, a < b)
	return result
}

func (c *CPU) sbc8(a, b byte) byte {
	carry := c.carryIn()
	result := a - b - carry
	c.regs.SetFlag(FlagZ, result == 0)
	c.regs.SetFlag(FlagN, true)
	c.regs.SetFlag(FlagH, int(a&0xF)-int(b&0xF)-int(carry) < 0)
	c.regs.SetFlag(FlagC, int(a)-int(b)-int(carry) < 0)
	return result
}

func (c *CPU) and8(a, b byte) byte {
	result := a & b
	c.regs.SetFlag(FlagZ, result == 0)
	c.regs.SetFlag(FlagN, false)
	c.regs.SetFlag(FlagH, true)
	c.regs.SetFlag(FlagC, false)
	return result
}

func (c *CPU) or8(a, b byte) byte {
	result := a | b
	c.regs.SetFlag(FlagZ, result == 0)
	c.regs.SetFlag(FlagN, false)
	c.regs.SetFlag(FlagH, false)
	c.regs.SetFlag(FlagC, false)
	return result
}

func (c *CPU) xor8(a, b byte) byte {
	result := a ^ b
	c.regs.SetFlag(FlagZ, result == 0)
	c.regs.SetFlag(FlagN, false)
	c.regs.SetFlag(FlagH, false)
	c.regs.SetFlag(FlagC, false)
	return result
}

func (c *CPU) cp8(a, b byte) {
	c.sub8(a, b)
}

func (c *CPU) inc8(v byte) byte {
	result := v + 1
	c.regs.SetFlag(FlagZ, result == 0)
	c.regs.SetFlag(FlagN, false)
	c.regs.SetFlag(FlagH, v&0xF == 0xF)
	return result
}

func (c *CPU) dec8(v byte) byte {
	result := v - 1
	c.regs.SetFlag(FlagZ, result == 0)
	c.regs.SetFlag(FlagN, true)
	c.regs.SetFlag(FlagH, v&0xF == 0)
	return result
}

func (c *CPU) carryIn() byte {
	if c.regs.HasFlag(FlagC) {
		return 1
	}
	return 0
}
