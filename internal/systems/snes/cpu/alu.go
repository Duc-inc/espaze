package cpu

// adc8/adc16/sbc8/sbc16 implement binary (non-decimal) add/subtract
// with carry - this project doesn't reproduce the 65816's BCD
// (Decimal flag) arithmetic mode, matching the pattern this project's
// other 6502-family cores already use (NES's 6502 has no decimal mode
// at all; the PC Engine's HuC6280 always runs ADC/SBC in binary
// regardless of the Decimal flag).
func (c *CPU) adc8(a, b byte) byte {
	carry := uint16(0)
	if c.regs.getFlag(FlagCarry) {
		carry = 1
	}
	sum := uint16(a) + uint16(b) + carry
	result := byte(sum)
	c.regs.setFlag(FlagCarry, sum > 0xFF)
	c.regs.setFlag(FlagOverflow, (a^result)&(b^result)&0x80 != 0)
	c.regs.setNZ8(result)
	return result
}

func (c *CPU) adc16(a, b uint16) uint16 {
	carry := uint32(0)
	if c.regs.getFlag(FlagCarry) {
		carry = 1
	}
	sum := uint32(a) + uint32(b) + carry
	result := uint16(sum)
	c.regs.setFlag(FlagCarry, sum > 0xFFFF)
	c.regs.setFlag(FlagOverflow, (a^result)&(b^result)&0x8000 != 0)
	c.regs.setNZ16(result)
	return result
}

func (c *CPU) sbc8(a, b byte) byte      { return c.adc8(a, ^b) }
func (c *CPU) sbc16(a, b uint16) uint16 { return c.adc16(a, ^b) }

func (c *CPU) cmp8(a, b byte) {
	result := a - b
	c.regs.setFlag(FlagCarry, a >= b)
	c.regs.setNZ8(result)
}

func (c *CPU) cmp16(a, b uint16) {
	result := a - b
	c.regs.setFlag(FlagCarry, a >= b)
	c.regs.setNZ16(result)
}

func (c *CPU) asl8(v byte) byte {
	c.regs.setFlag(FlagCarry, v&0x80 != 0)
	result := v << 1
	c.regs.setNZ8(result)
	return result
}

func (c *CPU) asl16(v uint16) uint16 {
	c.regs.setFlag(FlagCarry, v&0x8000 != 0)
	result := v << 1
	c.regs.setNZ16(result)
	return result
}

func (c *CPU) lsr8(v byte) byte {
	c.regs.setFlag(FlagCarry, v&0x01 != 0)
	result := v >> 1
	c.regs.setNZ8(result)
	return result
}

func (c *CPU) lsr16(v uint16) uint16 {
	c.regs.setFlag(FlagCarry, v&0x0001 != 0)
	result := v >> 1
	c.regs.setNZ16(result)
	return result
}

func (c *CPU) rol8(v byte) byte {
	carryIn := byte(0)
	if c.regs.getFlag(FlagCarry) {
		carryIn = 1
	}
	c.regs.setFlag(FlagCarry, v&0x80 != 0)
	result := v<<1 | carryIn
	c.regs.setNZ8(result)
	return result
}

func (c *CPU) rol16(v uint16) uint16 {
	carryIn := uint16(0)
	if c.regs.getFlag(FlagCarry) {
		carryIn = 1
	}
	c.regs.setFlag(FlagCarry, v&0x8000 != 0)
	result := v<<1 | carryIn
	c.regs.setNZ16(result)
	return result
}

func (c *CPU) ror8(v byte) byte {
	carryIn := byte(0)
	if c.regs.getFlag(FlagCarry) {
		carryIn = 0x80
	}
	c.regs.setFlag(FlagCarry, v&0x01 != 0)
	result := v>>1 | carryIn
	c.regs.setNZ8(result)
	return result
}

func (c *CPU) ror16(v uint16) uint16 {
	carryIn := uint16(0)
	if c.regs.getFlag(FlagCarry) {
		carryIn = 0x8000
	}
	c.regs.setFlag(FlagCarry, v&0x0001 != 0)
	result := v>>1 | carryIn
	c.regs.setNZ16(result)
	return result
}
