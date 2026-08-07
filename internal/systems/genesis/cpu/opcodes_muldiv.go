package cpu

// muldivOpcodes registers MULU/MULS (16x16 -> 32) and DIVU/DIVS (32/16
// -> 16-bit quotient + 16-bit remainder, packed into one Dn) - the
// 68000 has no 32-bit multiply or division wider than this.
func muldivOpcodes() []pattern {
	return []pattern{
		{mask: 0xF1C0, match: 0xC0C0, execute: muluExecute},
		{mask: 0xF1C0, match: 0xC1C0, execute: mulsExecute},
		{mask: 0xF1C0, match: 0x80C0, execute: divuExecute},
		{mask: 0xF1C0, match: 0x81C0, execute: divsExecute},
	}
}

func muluExecute(c *CPU, opcode uint16) int {
	reg := byte((opcode >> 9) & 0x07)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	src := uint32(c.readEA(mode, eaReg, sizeWord))
	dn := uint32(uint16(c.regs.D[reg]))
	result := src * dn

	c.regs.D[reg] = result
	c.regs.setFlag(FlagN, int32(result) < 0)
	c.regs.setFlag(FlagZ, result == 0)
	c.regs.setFlag(FlagV, false)
	c.regs.setFlag(FlagC, false)
	return 70 // real timing varies with operand value; this is a representative average
}

func mulsExecute(c *CPU, opcode uint16) int {
	reg := byte((opcode >> 9) & 0x07)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	src := int32(int16(c.readEA(mode, eaReg, sizeWord)))
	dn := int32(int16(c.regs.D[reg]))
	result := uint32(src * dn)

	c.regs.D[reg] = result
	c.regs.setFlag(FlagN, int32(result) < 0)
	c.regs.setFlag(FlagZ, result == 0)
	c.regs.setFlag(FlagV, false)
	c.regs.setFlag(FlagC, false)
	return 70
}

func divuExecute(c *CPU, opcode uint16) int {
	reg := byte((opcode >> 9) & 0x07)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	divisor := uint32(c.readEA(mode, eaReg, sizeWord))
	if divisor == 0 {
		return c.raiseException(vectorZeroDivide)
	}

	dividend := c.regs.D[reg]
	quotient := dividend / divisor
	if quotient > 0xFFFF { // overflow: quotient doesn't fit in 16 bits
		c.regs.setFlag(FlagV, true)
		return 10
	}
	remainder := dividend % divisor

	c.regs.D[reg] = remainder<<16 | (quotient & 0xFFFF)
	c.regs.setFlag(FlagN, int16(quotient) < 0)
	c.regs.setFlag(FlagZ, quotient == 0)
	c.regs.setFlag(FlagV, false)
	c.regs.setFlag(FlagC, false)
	return 140
}

func divsExecute(c *CPU, opcode uint16) int {
	reg := byte((opcode >> 9) & 0x07)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	divisor := int32(int16(c.readEA(mode, eaReg, sizeWord)))
	if divisor == 0 {
		return c.raiseException(vectorZeroDivide)
	}

	dividend := int32(c.regs.D[reg])
	quotient := dividend / divisor
	if quotient > 32767 || quotient < -32768 {
		c.regs.setFlag(FlagV, true)
		return 10
	}
	remainder := dividend % divisor

	c.regs.D[reg] = uint32(remainder)<<16 | uint32(uint16(quotient))
	c.regs.setFlag(FlagN, quotient < 0)
	c.regs.setFlag(FlagZ, quotient == 0)
	c.regs.setFlag(FlagV, false)
	c.regs.setFlag(FlagC, false)
	return 158
}
