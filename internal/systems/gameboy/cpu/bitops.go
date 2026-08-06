package cpu

// Rotate/shift/bit helpers used by both the CB-prefixed opcode table and
// the four legacy accumulator rotates (RLCA/RRCA/RLA/RRA), which reuse
// these but always force Z back to false afterward.

func (c *CPU) rlc(v byte) byte {
	carry := v&0x80 != 0
	result := v << 1
	if carry {
		result |= 0x01
	}
	c.setRotateFlags(result, carry)
	return result
}

func (c *CPU) rrc(v byte) byte {
	carry := v&0x01 != 0
	result := v >> 1
	if carry {
		result |= 0x80
	}
	c.setRotateFlags(result, carry)
	return result
}

func (c *CPU) rl(v byte) byte {
	result := v << 1
	if c.regs.HasFlag(FlagC) {
		result |= 0x01
	}
	c.setRotateFlags(result, v&0x80 != 0)
	return result
}

func (c *CPU) rr(v byte) byte {
	result := v >> 1
	if c.regs.HasFlag(FlagC) {
		result |= 0x80
	}
	c.setRotateFlags(result, v&0x01 != 0)
	return result
}

func (c *CPU) sla(v byte) byte {
	result := v << 1
	c.setRotateFlags(result, v&0x80 != 0)
	return result
}

func (c *CPU) sra(v byte) byte {
	result := (v >> 1) | (v & 0x80)
	c.setRotateFlags(result, v&0x01 != 0)
	return result
}

func (c *CPU) srl(v byte) byte {
	result := v >> 1
	c.setRotateFlags(result, v&0x01 != 0)
	return result
}

func (c *CPU) swap(v byte) byte {
	result := v<<4 | v>>4
	c.regs.SetFlag(FlagZ, result == 0)
	c.regs.SetFlag(FlagN, false)
	c.regs.SetFlag(FlagH, false)
	c.regs.SetFlag(FlagC, false)
	return result
}

func (c *CPU) setRotateFlags(result byte, carry bool) {
	c.regs.SetFlag(FlagZ, result == 0)
	c.regs.SetFlag(FlagN, false)
	c.regs.SetFlag(FlagH, false)
	c.regs.SetFlag(FlagC, carry)
}

// bit tests bit n of v (BIT instruction): Z reflects the bit, C untouched.
func (c *CPU) bit(v byte, n byte) {
	c.regs.SetFlag(FlagZ, v&(1<<n) == 0)
	c.regs.SetFlag(FlagN, false)
	c.regs.SetFlag(FlagH, true)
}
