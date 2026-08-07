package cpu

// decodeIndexed handles the DD/FD-prefixed opcode space: every regular
// instruction that references HL/(HL) gets IX/IY /(IX+d) or (IY+d)
// substituted instead. Real hardware falls back to executing the very
// same opcode as if the prefix hadn't been there (just 4 cycles slower)
// for any instruction that *doesn't* touch H/L/(HL); rather than
// duplicating the entire main decoder to reproduce every one of those
// pass-through cases, this implements the common IX/IY-specific forms
// explicitly and falls back to the plain decoder - which is exactly
// correct - for everything else.
func (c *CPU) decodeIndexed(ix *uint16) int {
	op := c.fetchByte()

	switch op {
	case 0x21: // LD IX,nn
		*ix = c.fetch16()
		return 14
	case 0x22: // LD (nn),IX
		c.write16(c.fetch16(), *ix)
		return 20
	case 0x2A: // LD IX,(nn)
		*ix = c.read16(c.fetch16())
		return 20
	case 0x23: // INC IX
		*ix++
		return 10
	case 0x2B: // DEC IX
		*ix--
		return 10
	case 0x09, 0x19, 0x29, 0x39: // ADD IX,rp
		*ix = c.addHL16(*ix, c.indexedRP(op, ix))
		return 15
	case 0x34, 0x35, 0x36: // INC/DEC/LD (IX+d),n
		return c.indexedRMW8(op, ix)
	case 0xE1: // POP IX
		*ix = c.pop()
		return 14
	case 0xE5: // PUSH IX
		c.push(*ix)
		return 15
	case 0xE3: // EX (SP),IX
		v := c.read16(c.regs.SP)
		c.write16(c.regs.SP, *ix)
		*ix = v
		return 23
	case 0xE9: // JP (IX)
		c.regs.PC = *ix
		return 8
	case 0xF9: // LD SP,IX
		c.regs.SP = *ix
		return 10
	case 0xCB: // DDCB/FDCB: displacement byte, then a rotate/bit/set/res opcode
		return c.decodeIndexedCB(ix)
	}

	if op >= 0x40 && op <= 0xBF && op != 0x76 {
		if r := c.indexedLoadOrALU(op, ix); r != 0 {
			return r
		}
	}

	// Doesn't touch H/L/(HL): identical to the unprefixed instruction,
	// just charged the prefix fetch's extra 4 cycles.
	c.regs.PC--
	return 4 + c.decode()
}

// indexedRP mirrors rp(), except slot 2 (normally HL) means *ix here.
func (c *CPU) indexedRP(op byte, ix *uint16) uint16 {
	switch (op >> 4) & 0x03 {
	case 0:
		return c.regs.BC()
	case 1:
		return c.regs.DE()
	case 2:
		return *ix
	default:
		return c.regs.SP
	}
}

func (c *CPU) indexedRMW8(op byte, ix *uint16) int {
	d := int8(c.fetchByte())
	addr := uint16(int32(*ix) + int32(d))
	switch op {
	case 0x34:
		c.bus.Write(addr, c.inc8(c.bus.Read(addr)))
		return 23
	case 0x35:
		c.bus.Write(addr, c.dec8(c.bus.Read(addr)))
		return 23
	default: // 0x36
		n := c.fetchByte()
		c.bus.Write(addr, n)
		return 19
	}
}

// indexedLoadOrALU covers LD r,(IX+d) / LD (IX+d),r / ALU A,(IX+d) - the
// bulk of real-world IX/IY use - by re-running the normal x=1/x=2 field
// decode with r8(6)/setR8(6) redirected to (IX+d) instead of (HL).
// Returns 0 (meaning "not one of these, try elsewhere") for anything
// that doesn't actually reference index 6 on either side.
func (c *CPU) indexedLoadOrALU(op byte, ix *uint16) int {
	x, y, z, _, _ := decomposeOpcode(op)
	if z != 6 && y != 6 {
		return 0 // neither operand is (HL) - not something IX/IY changes
	}

	d := int8(c.fetchByte())
	addr := uint16(int32(*ix) + int32(d))

	switch x {
	case 1: // LD
		if z == 6 {
			c.setR8(y, c.bus.Read(addr))
		} else {
			c.bus.Write(addr, c.r8(z))
		}
		return 19
	default: // x == 2, ALU A,(IX+d)
		c.aluOp(y, c.bus.Read(addr))
		return 19
	}
}

// decodeIndexedCB handles DD CB d op / FD CB d op: rotate/shift/bit/
// set/res, always against (IX+d)/(IY+d) regardless of the z field
// (real hardware's undocumented "also store into a register" variants
// aren't implemented - just the documented (IX+d)-only behavior).
func (c *CPU) decodeIndexedCB(ix *uint16) int {
	d := int8(c.fetchByte())
	op := c.fetchByte()
	addr := uint16(int32(*ix) + int32(d))
	x, y, _, _, _ := decomposeOpcode(op)

	v := c.bus.Read(addr)
	switch x {
	case 0:
		c.bus.Write(addr, c.rotOp(y, v))
		return 23
	case 1:
		c.bitOp(y, v)
		return 20
	case 2:
		c.bus.Write(addr, v&^(1<<y))
		return 23
	default:
		c.bus.Write(addr, v|(1<<y))
		return 23
	}
}
