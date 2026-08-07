package spc700

func (c *CPU) addWithCarry(a, b byte, carryIn bool) byte {
	carry := uint16(0)
	if carryIn {
		carry = 1
	}
	sum := uint16(a) + uint16(b) + carry
	result := byte(sum)
	c.regs.setFlag(FlagCarry, sum > 0xFF)
	c.regs.setFlag(FlagOverflow, (a^result)&(b^result)&0x80 != 0)
	c.regs.setNZ(result)
	return result
}

func (c *CPU) adc(a, b byte) byte { return c.addWithCarry(a, b, c.regs.getFlag(FlagCarry)) }
func (c *CPU) add(a, b byte) byte { return c.addWithCarry(a, b, false) }
func (c *CPU) sbc(a, b byte) byte { return c.addWithCarry(a, ^b, c.regs.getFlag(FlagCarry)) }
func (c *CPU) sub(a, b byte) byte { return c.addWithCarry(a, ^b, true) }

func (c *CPU) cmp(a, b byte) {
	result := a - b
	c.regs.setFlag(FlagCarry, a >= b)
	c.regs.setNZ(result)
}

// applyALU dispatches one of this project's 8 accumulator ALU
// operations by index, matching opcodes_alu.go's own numbering.
func (c *CPU) applyALU(op byte, a, b byte) byte {
	switch op {
	case 0:
		return c.add(a, b)
	case 1:
		return c.adc(a, b)
	case 2:
		return c.sub(a, b)
	case 3:
		return c.sbc(a, b)
	case 4:
		c.cmp(a, b)
		return a
	case 5:
		r := a & b
		c.regs.setNZ(r)
		return r
	case 6:
		r := a | b
		c.regs.setNZ(r)
		return r
	default:
		r := a ^ b
		c.regs.setNZ(r)
		return r
	}
}
