package cpu

// aluOp performs one of the 8 accumulator ALU operations the y field
// selects (ADD, ADC, SUB, SBC, AND, XOR, OR, CP) against operand,
// updating A (except CP, which only sets flags) and every flag exactly
// as real hardware does - including the undocumented Y/X flags, which
// CP copies from the operand rather than the result, a real quirk.
func (c *CPU) aluOp(op byte, operand byte) {
	a := c.regs.A
	switch op {
	case 0: // ADD
		c.add8(operand, false)
	case 1: // ADC
		c.add8(operand, c.regs.getFlag(FlagC))
	case 2: // SUB
		c.sub8(operand, false, true)
	case 3: // SBC
		c.sub8(operand, c.regs.getFlag(FlagC), true)
	case 4: // AND
		c.regs.A &= operand
		c.regs.setSZ(c.regs.A)
		c.regs.setYX(c.regs.A)
		c.regs.setFlag(FlagH, true)
		c.regs.setFlag(FlagPV, parity(c.regs.A))
		c.regs.setFlag(FlagN, false)
		c.regs.setFlag(FlagC, false)
	case 5: // XOR
		c.regs.A ^= operand
		c.regs.setSZ(c.regs.A)
		c.regs.setYX(c.regs.A)
		c.regs.setFlag(FlagH, false)
		c.regs.setFlag(FlagPV, parity(c.regs.A))
		c.regs.setFlag(FlagN, false)
		c.regs.setFlag(FlagC, false)
	case 6: // OR
		c.regs.A |= operand
		c.regs.setSZ(c.regs.A)
		c.regs.setYX(c.regs.A)
		c.regs.setFlag(FlagH, false)
		c.regs.setFlag(FlagPV, parity(c.regs.A))
		c.regs.setFlag(FlagN, false)
		c.regs.setFlag(FlagC, false)
	default: // CP: like SUB, but discards the result and takes Y/X from the operand
		c.sub8(operand, false, false)
		c.regs.setYX(operand)
		c.regs.A = a
	}
}

// add8 implements ADD/ADC A,operand.
func (c *CPU) add8(operand byte, carryIn bool) {
	a := c.regs.A
	cin := byte(0)
	if carryIn {
		cin = 1
	}
	sum := uint16(a) + uint16(operand) + uint16(cin)
	result := byte(sum)

	c.regs.setFlag(FlagC, sum > 0xFF)
	c.regs.setFlag(FlagH, (a&0x0F)+(operand&0x0F)+cin > 0x0F)
	c.regs.setFlag(FlagPV, (a^operand)&0x80 == 0 && (a^result)&0x80 != 0)
	c.regs.setFlag(FlagN, false)
	c.regs.setSZ(result)
	c.regs.setYX(result)
	c.regs.A = result
}

// sub8 implements SUB/SBC/CP A,operand (storeResult false is used
// internally by CP's flag-only pass).
func (c *CPU) sub8(operand byte, carryIn bool, storeResult bool) {
	a := c.regs.A
	cin := byte(0)
	if carryIn {
		cin = 1
	}
	diff := int16(a) - int16(operand) - int16(cin)
	result := byte(diff)

	c.regs.setFlag(FlagC, diff < 0)
	c.regs.setFlag(FlagH, int16(a&0x0F)-int16(operand&0x0F)-int16(cin) < 0)
	c.regs.setFlag(FlagPV, (a^operand)&0x80 != 0 && (a^result)&0x80 != 0)
	c.regs.setFlag(FlagN, true)
	c.regs.setSZ(result)
	c.regs.setYX(result)
	if storeResult {
		c.regs.A = result
	}
}

// inc8/dec8 implement INC r/(HL) and DEC r/(HL) - unlike SUB/ADD, these
// never touch the Carry flag.
func (c *CPU) inc8(v byte) byte {
	result := v + 1
	c.regs.setFlag(FlagH, v&0x0F == 0x0F)
	c.regs.setFlag(FlagPV, v == 0x7F)
	c.regs.setFlag(FlagN, false)
	c.regs.setSZ(result)
	c.regs.setYX(result)
	return result
}

func (c *CPU) dec8(v byte) byte {
	result := v - 1
	c.regs.setFlag(FlagH, v&0x0F == 0x00)
	c.regs.setFlag(FlagPV, v == 0x80)
	c.regs.setFlag(FlagN, true)
	c.regs.setSZ(result)
	c.regs.setYX(result)
	return result
}

// addHL16 implements ADD HL,rp - unlike the 8-bit ALU ops it leaves
// S/Z/PV alone and only updates H/N/C plus the undocumented flags (from
// the result's high byte).
func (c *CPU) addHL16(hl, operand uint16) uint16 {
	sum := uint32(hl) + uint32(operand)
	result := uint16(sum)

	c.regs.setFlag(FlagC, sum > 0xFFFF)
	c.regs.setFlag(FlagH, (hl&0x0FFF)+(operand&0x0FFF) > 0x0FFF)
	c.regs.setFlag(FlagN, false)
	c.regs.setYX(byte(result >> 8))
	return result
}
