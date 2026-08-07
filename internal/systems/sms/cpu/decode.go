package cpu

// decode fetches and executes exactly one unprefixed instruction (or
// routes to a prefixed group), returning its T-state cost. See
// regcodes.go for the x/y/z/p/q field decomposition this and the other
// opcodes_*.go files dispatch on.
func (c *CPU) decode() int {
	op := c.fetchByte()
	x, y, z, p, q := decomposeOpcode(op)

	switch x {
	case 0:
		return c.decodeX0(y, z, p, q)
	case 1:
		return c.decodeX1(y, z)
	case 2:
		return c.decodeX2(y, z)
	default:
		return c.decodeX3(y, z, p, q)
	}
}

func (c *CPU) decodeX0(y, z, p, q byte) int {
	switch z {
	case 0:
		return c.decodeX0Z0(y)
	case 1:
		if q == 0 {
			c.setRP(p, c.fetch16())
			return 10
		}
		c.regs.SetHL(c.addHL16(c.regs.HL(), c.rp(p)))
		return 11
	case 2:
		return c.decodeX0Z2(p, q)
	case 3:
		if q == 0 {
			c.setRP(p, c.rp(p)+1)
		} else {
			c.setRP(p, c.rp(p)-1)
		}
		return 6
	case 4:
		return c.incDecReg(y, true)
	case 5:
		return c.incDecReg(y, false)
	case 6:
		n := c.fetchByte()
		c.setR8(y, n)
		if y == 6 {
			return 10
		}
		return 7
	default: // z == 7
		return c.decodeX0Z7(y)
	}
}

func (c *CPU) decodeX0Z0(y byte) int {
	switch y {
	case 0: // NOP
		return 4
	case 1: // EX AF,AF'
		c.regs.A, c.regs.A2 = c.regs.A2, c.regs.A
		c.regs.F, c.regs.F2 = c.regs.F2, c.regs.F
		return 4
	case 2: // DJNZ d
		d := int8(c.fetchByte())
		c.regs.B--
		if c.regs.B != 0 {
			c.regs.PC = uint16(int32(c.regs.PC) + int32(d))
			return 13
		}
		return 8
	case 3: // JR d
		d := int8(c.fetchByte())
		c.regs.PC = uint16(int32(c.regs.PC) + int32(d))
		return 12
	default: // JR cc,d (y = 4..7 -> condition 0..3)
		d := int8(c.fetchByte())
		if c.condition(y - 4) {
			c.regs.PC = uint16(int32(c.regs.PC) + int32(d))
			return 12
		}
		return 7
	}
}

func (c *CPU) decodeX0Z2(p, q byte) int {
	if q == 0 {
		switch p {
		case 0:
			c.bus.Write(c.regs.BC(), c.regs.A)
		case 1:
			c.bus.Write(c.regs.DE(), c.regs.A)
		case 2:
			c.write16(c.fetch16(), c.regs.HL())
		default:
			c.bus.Write(c.fetch16(), c.regs.A)
		}
	} else {
		switch p {
		case 0:
			c.regs.A = c.bus.Read(c.regs.BC())
		case 1:
			c.regs.A = c.bus.Read(c.regs.DE())
		case 2:
			c.regs.SetHL(c.read16(c.fetch16()))
		default:
			c.regs.A = c.bus.Read(c.fetch16())
		}
	}
	if p >= 2 {
		if q == 0 && p == 3 {
			return 13
		}
		if q == 1 && p == 3 {
			return 13
		}
		return 16
	}
	return 7
}

func (c *CPU) incDecReg(y byte, inc bool) int {
	v := c.r8(y)
	if inc {
		c.setR8(y, c.inc8(v))
	} else {
		c.setR8(y, c.dec8(v))
	}
	if y == 6 {
		return 11
	}
	return 4
}

func (c *CPU) decodeX0Z7(y byte) int {
	switch y {
	case 0: // RLCA
		carry := c.regs.A&0x80 != 0
		c.regs.A = c.regs.A<<1 | b2u8(carry)
		c.regs.setFlag(FlagC, carry)
		c.regs.setFlag(FlagH, false)
		c.regs.setFlag(FlagN, false)
		c.regs.setYX(c.regs.A)
	case 1: // RRCA
		carry := c.regs.A&0x01 != 0
		c.regs.A = c.regs.A>>1 | b2u8(carry)<<7
		c.regs.setFlag(FlagC, carry)
		c.regs.setFlag(FlagH, false)
		c.regs.setFlag(FlagN, false)
		c.regs.setYX(c.regs.A)
	case 2: // RLA
		carry := c.regs.A&0x80 != 0
		c.regs.A = c.regs.A<<1 | b2u8(c.regs.getFlag(FlagC))
		c.regs.setFlag(FlagC, carry)
		c.regs.setFlag(FlagH, false)
		c.regs.setFlag(FlagN, false)
		c.regs.setYX(c.regs.A)
	case 3: // RRA
		carry := c.regs.A&0x01 != 0
		c.regs.A = c.regs.A>>1 | b2u8(c.regs.getFlag(FlagC))<<7
		c.regs.setFlag(FlagC, carry)
		c.regs.setFlag(FlagH, false)
		c.regs.setFlag(FlagN, false)
		c.regs.setYX(c.regs.A)
	case 4: // DAA
		c.daa()
	case 5: // CPL
		c.regs.A = ^c.regs.A
		c.regs.setFlag(FlagH, true)
		c.regs.setFlag(FlagN, true)
		c.regs.setYX(c.regs.A)
	case 6: // SCF
		c.regs.setFlag(FlagC, true)
		c.regs.setFlag(FlagH, false)
		c.regs.setFlag(FlagN, false)
		c.regs.setYX(c.regs.A)
	default: // CCF
		c.regs.setFlag(FlagH, c.regs.getFlag(FlagC))
		c.regs.setFlag(FlagC, !c.regs.getFlag(FlagC))
		c.regs.setFlag(FlagN, false)
		c.regs.setYX(c.regs.A)
	}
	return 4
}

func b2u8(b bool) byte {
	if b {
		return 1
	}
	return 0
}

// daa implements DAA's famously fiddly BCD-correction table.
func (c *CPU) daa() {
	a := c.regs.A
	adjust := byte(0)
	carry := c.regs.getFlag(FlagC)

	if c.regs.getFlag(FlagH) || a&0x0F > 9 {
		adjust |= 0x06
	}
	if carry || a > 0x99 {
		adjust |= 0x60
		carry = true
	}

	if c.regs.getFlag(FlagN) {
		a -= adjust
	} else {
		a += adjust
	}

	c.regs.setSZ(a)
	c.regs.setYX(a)
	c.regs.setFlag(FlagPV, parity(a))
	c.regs.setFlag(FlagC, carry)
	c.regs.setFlag(FlagH, false) // exact half-carry-of-DAA rules are notoriously undocumented; this matches the commonly cited approximation
	c.regs.A = a
}
