package cpu

func (c *CPU) add8(a, b byte, carryIn bool) byte {
	carry := byte(0)
	if carryIn {
		carry = 1
	}
	sum := uint16(a) + uint16(b) + uint16(carry)
	result := byte(sum)
	c.regs.setFlag(FlagCarry, sum > 0xFF)
	c.regs.setFlag(FlagHalfCarry, (a&0x0F)+(b&0x0F)+carry > 0x0F)
	c.regs.setFlag(FlagOverflow, (a^result)&(b^result)&0x80 != 0)
	c.regs.setFlag(FlagNegative, false)
	c.regs.setSZ8(result)
	return result
}

func (c *CPU) sub8(a, b byte, borrowIn bool) byte {
	borrow := byte(0)
	if borrowIn {
		borrow = 1
	}
	diff := int16(a) - int16(b) - int16(borrow)
	result := byte(diff)
	c.regs.setFlag(FlagCarry, diff < 0)
	c.regs.setFlag(FlagHalfCarry, int16(a&0x0F)-int16(b&0x0F)-int16(borrow) < 0)
	c.regs.setFlag(FlagOverflow, (a^b)&(a^result)&0x80 != 0)
	c.regs.setFlag(FlagNegative, true)
	c.regs.setSZ8(result)
	return result
}

func (c *CPU) add16(a, b uint16, carryIn bool) uint16 {
	carry := uint32(0)
	if carryIn {
		carry = 1
	}
	sum := uint32(a) + uint32(b) + carry
	result := uint16(sum)
	c.regs.setFlag(FlagCarry, sum > 0xFFFF)
	c.regs.setFlag(FlagOverflow, (a^result)&(b^result)&0x8000 != 0)
	c.regs.setFlag(FlagNegative, false)
	c.regs.setSZ16(result)
	return result
}

func (c *CPU) sub16(a, b uint16, borrowIn bool) uint16 {
	borrow := int32(0)
	if borrowIn {
		borrow = 1
	}
	diff := int32(a) - int32(b) - borrow
	result := uint16(diff)
	c.regs.setFlag(FlagCarry, diff < 0)
	c.regs.setFlag(FlagOverflow, (a^b)&(a^result)&0x8000 != 0)
	c.regs.setFlag(FlagNegative, true)
	c.regs.setSZ16(result)
	return result
}

func (c *CPU) applyALU8(op byte, a, b byte) byte {
	switch op {
	case 0:
		return c.add8(a, b, false)
	case 1:
		return c.add8(a, b, c.regs.getFlag(FlagCarry))
	case 2:
		return c.sub8(a, b, false)
	case 3:
		return c.sub8(a, b, c.regs.getFlag(FlagCarry))
	case 4:
		r := a & b
		c.regs.setSZ8(r)
		c.regs.setFlag(FlagCarry, false)
		return r
	case 5:
		r := a ^ b
		c.regs.setSZ8(r)
		c.regs.setFlag(FlagCarry, false)
		return r
	case 6:
		r := a | b
		c.regs.setSZ8(r)
		c.regs.setFlag(FlagCarry, false)
		return r
	default: // CP: like SUB but discards the result
		c.sub8(a, b, false)
		return a
	}
}

func (c *CPU) applyALU16(op byte, a, b uint16) uint16 {
	switch op {
	case 0:
		return c.add16(a, b, false)
	case 1:
		return c.add16(a, b, c.regs.getFlag(FlagCarry))
	case 2:
		return c.sub16(a, b, false)
	case 3:
		return c.sub16(a, b, c.regs.getFlag(FlagCarry))
	case 4:
		r := a & b
		c.regs.setSZ16(r)
		c.regs.setFlag(FlagCarry, false)
		return r
	case 5:
		r := a ^ b
		c.regs.setSZ16(r)
		c.regs.setFlag(FlagCarry, false)
		return r
	case 6:
		r := a | b
		c.regs.setSZ16(r)
		c.regs.setFlag(FlagCarry, false)
		return r
	default: // CP
		c.sub16(a, b, false)
		return a
	}
}
