package cpu

// decodeCB handles the CB-prefixed opcode space: rotate/shift (x=0),
// BIT (x=1), RES (x=2) and SET (x=3), each against any of the 8
// register/memory operands z selects.
func (c *CPU) decodeCB() int {
	op := c.fetchByte()
	x, y, z, _, _ := decomposeOpcode(op)

	v := c.r8(z)
	switch x {
	case 0:
		result := c.rotOp(y, v)
		c.setR8(z, result)
	case 1:
		c.bitOp(y, v)
	case 2:
		c.setR8(z, v&^(1<<y))
	default:
		c.setR8(z, v|(1<<y))
	}

	if z == 6 {
		if x == 1 {
			return 12
		}
		return 15
	}
	return 8
}

// rotOp implements the 8 CB rotate/shift variants (RLC, RRC, RL, RR,
// SLA, SRA, SLL - undocumented, sets bit 0 - and SRL), all sharing the
// same flag pattern: S/Z/PV(parity)/Y/X from the result, H and N clear,
// C from the bit shifted out.
func (c *CPU) rotOp(y byte, v byte) byte {
	var result byte
	var carry bool

	switch y {
	case 0: // RLC
		carry = v&0x80 != 0
		result = v<<1 | b2u8(carry)
	case 1: // RRC
		carry = v&0x01 != 0
		result = v>>1 | b2u8(carry)<<7
	case 2: // RL
		carry = v&0x80 != 0
		result = v<<1 | b2u8(c.regs.getFlag(FlagC))
	case 3: // RR
		carry = v&0x01 != 0
		result = v>>1 | b2u8(c.regs.getFlag(FlagC))<<7
	case 4: // SLA
		carry = v&0x80 != 0
		result = v << 1
	case 5: // SRA
		carry = v&0x01 != 0
		result = v>>1 | v&0x80
	case 6: // SLL (undocumented)
		carry = v&0x80 != 0
		result = v<<1 | 1
	default: // SRL
		carry = v&0x01 != 0
		result = v >> 1
	}

	c.regs.setSZ(result)
	c.regs.setYX(result)
	c.regs.setFlag(FlagPV, parity(result))
	c.regs.setFlag(FlagH, false)
	c.regs.setFlag(FlagN, false)
	c.regs.setFlag(FlagC, carry)
	return result
}

// bitOp implements BIT y,v: Z set if the bit is clear, H always set,
// N always clear, and (a real hardware quirk) S/PV mirror the tested
// bit when it's bit 7, while Y/X come from the operand rather than any
// computed result.
func (c *CPU) bitOp(y byte, v byte) {
	set := v&(1<<y) != 0
	c.regs.setFlag(FlagZ, !set)
	c.regs.setFlag(FlagPV, !set)
	c.regs.setFlag(FlagH, true)
	c.regs.setFlag(FlagN, false)
	c.regs.setFlag(FlagS, y == 7 && set)
	c.regs.setYX(v)
}
