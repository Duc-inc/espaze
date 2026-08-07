package cpu

// decodeED handles the ED-prefixed opcode space: 16-bit ADC/SBC, 16-bit
// loads to/from BC/DE/SP, block transfer/compare/IO instructions,
// interrupt mode/register access, and a handful of one-off instructions
// (NEG, RETN/RETI, RRD/RLD). Undefined ED opcodes act as an 8-cycle NOP
// on real hardware.
func (c *CPU) decodeED() int {
	op := c.fetchByte()
	x, y, z, p, q := decomposeOpcode(op)

	switch {
	case x == 1:
		return c.decodeEDX1(y, z, p, q)
	case x == 2 && y >= 4 && z <= 3:
		return c.decodeEDBlock(y, z)
	default:
		return 8 // undefined ED opcode
	}
}

func (c *CPU) decodeEDX1(y, z, p, q byte) int {
	switch z {
	case 0: // IN r,(C) / IN (C) (flags-only, no destination)
		v := c.io.In(c.regs.C)
		if y != 6 {
			c.setR8(y, v)
		}
		c.regs.setSZ(v)
		c.regs.setYX(v)
		c.regs.setFlag(FlagPV, parity(v))
		c.regs.setFlag(FlagH, false)
		c.regs.setFlag(FlagN, false)
		return 12
	case 1: // OUT (C),r / OUT (C),0
		v := byte(0)
		if y != 6 {
			v = c.r8(y)
		}
		c.io.Out(c.regs.C, v)
		return 12
	case 2:
		if q == 0 {
			c.regs.SetHL(c.sbcHL(c.regs.HL(), c.rp(p)))
		} else {
			c.regs.SetHL(c.adcHL(c.regs.HL(), c.rp(p)))
		}
		return 15
	case 3:
		addr := c.fetch16()
		if q == 0 {
			c.write16(addr, c.rp(p))
		} else {
			c.setRP(p, c.read16(addr))
		}
		return 20
	case 4: // NEG
		v := c.regs.A
		c.regs.A = 0
		c.sub8(v, false, true)
		return 8
	case 5: // RETN / RETI
		c.regs.IFF1 = c.regs.IFF2
		c.regs.PC = c.pop()
		return 14
	case 6: // IM 0/1/2
		c.regs.IM = [8]byte{0, 0, 1, 2, 0, 0, 1, 2}[y]
		return 8
	default:
		return c.decodeEDMisc(y)
	}
}

func (c *CPU) decodeEDMisc(y byte) int {
	switch y {
	case 0: // LD I,A
		c.regs.I = c.regs.A
		return 9
	case 1: // LD R,A
		c.regs.R = c.regs.A
		return 9
	case 2: // LD A,I
		c.regs.A = c.regs.I
		c.regs.setSZ(c.regs.A)
		c.regs.setYX(c.regs.A)
		c.regs.setFlag(FlagH, false)
		c.regs.setFlag(FlagN, false)
		c.regs.setFlag(FlagPV, c.regs.IFF2)
		return 9
	case 3: // LD A,R
		c.regs.A = c.regs.R
		c.regs.setSZ(c.regs.A)
		c.regs.setYX(c.regs.A)
		c.regs.setFlag(FlagH, false)
		c.regs.setFlag(FlagN, false)
		c.regs.setFlag(FlagPV, c.regs.IFF2)
		return 9
	case 4: // RRD
		c.rrd()
		return 18
	case 5: // RLD
		c.rld()
		return 18
	default: // undefined, acts as NOP
		return 8
	}
}

func (c *CPU) adcHL(hl, operand uint16) uint16 {
	cin := uint32(0)
	if c.regs.getFlag(FlagC) {
		cin = 1
	}
	sum := uint32(hl) + uint32(operand) + cin
	result := uint16(sum)

	c.regs.setFlag(FlagC, sum > 0xFFFF)
	c.regs.setFlag(FlagH, (hl&0x0FFF)+(operand&0x0FFF)+uint16(cin) > 0x0FFF)
	c.regs.setFlag(FlagPV, (hl^operand)&0x8000 == 0 && (hl^result)&0x8000 != 0)
	c.regs.setFlag(FlagN, false)
	c.regs.setSZ(byte(result >> 8))
	c.regs.setFlag(FlagZ, result == 0)
	c.regs.setYX(byte(result >> 8))
	return result
}

func (c *CPU) sbcHL(hl, operand uint16) uint16 {
	cin := int32(0)
	if c.regs.getFlag(FlagC) {
		cin = 1
	}
	diff := int32(hl) - int32(operand) - cin
	result := uint16(diff)

	c.regs.setFlag(FlagC, diff < 0)
	c.regs.setFlag(FlagH, int32(hl&0x0FFF)-int32(operand&0x0FFF)-cin < 0)
	c.regs.setFlag(FlagPV, (hl^operand)&0x8000 != 0 && (hl^result)&0x8000 != 0)
	c.regs.setFlag(FlagN, true)
	c.regs.setSZ(byte(result >> 8))
	c.regs.setFlag(FlagZ, result == 0)
	c.regs.setYX(byte(result >> 8))
	return result
}

// rrd/rld rotate a nibble between A and (HL) - real hardware's
// nibble-swap instructions, occasionally used for BCD digit shuffling.
func (c *CPU) rrd() {
	hl := c.regs.HL()
	m := c.bus.Read(hl)
	newA := c.regs.A&0xF0 | m&0x0F
	newM := c.regs.A<<4 | m>>4
	c.regs.A = newA
	c.bus.Write(hl, newM)
	c.regs.setSZ(c.regs.A)
	c.regs.setYX(c.regs.A)
	c.regs.setFlag(FlagH, false)
	c.regs.setFlag(FlagPV, parity(c.regs.A))
	c.regs.setFlag(FlagN, false)
}

func (c *CPU) rld() {
	hl := c.regs.HL()
	m := c.bus.Read(hl)
	newA := c.regs.A&0xF0 | m>>4
	newM := m<<4 | c.regs.A&0x0F
	c.regs.A = newA
	c.bus.Write(hl, newM)
	c.regs.setSZ(c.regs.A)
	c.regs.setYX(c.regs.A)
	c.regs.setFlag(FlagH, false)
	c.regs.setFlag(FlagPV, parity(c.regs.A))
	c.regs.setFlag(FlagN, false)
}
