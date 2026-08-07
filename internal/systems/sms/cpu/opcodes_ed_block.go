package cpu

// decodeEDBlock handles the 16 block transfer/compare/IO instructions
// (LDI/LDD/LDIR/LDDR, CPI/CPD/CPIR/CPDR, INI/IND/INIR/INDR,
// OUTI/OUTD/OTIR/OTDR) - z selects the operation, y selects
// increment-vs-decrement and single-shot-vs-repeating.
func (c *CPU) decodeEDBlock(y, z byte) int {
	decrement := y == 5 || y == 7
	repeat := y == 6 || y == 7

	switch z {
	case 0:
		return c.blockLD(decrement, repeat)
	case 1:
		return c.blockCP(decrement, repeat)
	case 2:
		return c.blockIN(decrement, repeat)
	default:
		return c.blockOUT(decrement, repeat)
	}
}

func step16(v *uint16, decrement bool) {
	if decrement {
		*v--
	} else {
		*v++
	}
}

// blockLD implements LDI/LDD/LDIR/LDDR: copy (HL) to (DE), step both,
// decrement BC. The undocumented Y/X flags come from (transferred byte
// + A), a real quirk some emulator test suites check for.
func (c *CPU) blockLD(decrement, repeat bool) int {
	hl, de := c.regs.HL(), c.regs.DE()
	v := c.bus.Read(hl)
	c.bus.Write(de, v)
	step16(&hl, decrement)
	step16(&de, decrement)
	c.regs.SetHL(hl)
	c.regs.SetDE(de)

	bc := c.regs.BC() - 1
	c.regs.SetBC(bc)

	n := v + c.regs.A
	c.regs.setFlag(FlagY, n&0x02 != 0)
	c.regs.setFlag(FlagX, n&0x08 != 0)
	c.regs.setFlag(FlagH, false)
	c.regs.setFlag(FlagN, false)
	c.regs.setFlag(FlagPV, bc != 0)

	if repeat && bc != 0 {
		c.regs.PC -= 2
		return 21
	}
	return 16
}

// blockCP implements CPI/CPD/CPIR/CPDR: compare A with (HL) like CP,
// but only steps HL and decrements BC rather than touching A.
func (c *CPU) blockCP(decrement, repeat bool) int {
	hl := c.regs.HL()
	v := c.bus.Read(hl)
	step16(&hl, decrement)
	c.regs.SetHL(hl)

	bc := c.regs.BC() - 1
	c.regs.SetBC(bc)

	a := c.regs.A
	result := a - v
	halfBorrow := a&0x0F < v&0x0F

	c.regs.setSZ(result)
	c.regs.setFlag(FlagH, halfBorrow)
	c.regs.setFlag(FlagPV, bc != 0)
	c.regs.setFlag(FlagN, true)

	n := result
	if halfBorrow {
		n--
	}
	c.regs.setFlag(FlagY, n&0x02 != 0)
	c.regs.setFlag(FlagX, n&0x08 != 0)

	if repeat && bc != 0 && result != 0 {
		c.regs.PC -= 2
		return 21
	}
	return 16
}

// blockIN/blockOUT implement INI/IND/INIR/INDR/OUTI/OUTD/OTIR/OTDR.
// Real hardware's flags for these involve the I/O port's own carry-out
// behavior in a way that's rarely relied on by software; this sets the
// commonly-checked ones (Z from B, N from the transferred byte's sign)
// and leaves the more obscure H/C/PV rules at their conventional
// approximation rather than hardware-exact.
func (c *CPU) blockIN(decrement, repeat bool) int {
	v := c.io.In(c.regs.C)
	hl := c.regs.HL()
	c.bus.Write(hl, v)
	step16(&hl, decrement)
	c.regs.SetHL(hl)

	c.regs.B--
	c.regs.setFlag(FlagZ, c.regs.B == 0)
	c.regs.setFlag(FlagN, v&0x80 != 0)

	if repeat && c.regs.B != 0 {
		c.regs.PC -= 2
		return 21
	}
	return 16
}

func (c *CPU) blockOUT(decrement, repeat bool) int {
	hl := c.regs.HL()
	v := c.bus.Read(hl)
	c.io.Out(c.regs.C, v)
	step16(&hl, decrement)
	c.regs.SetHL(hl)

	c.regs.B--
	c.regs.setFlag(FlagZ, c.regs.B == 0)
	c.regs.setFlag(FlagN, v&0x80 != 0)

	if repeat && c.regs.B != 0 {
		c.regs.PC -= 2
		return 21
	}
	return 16
}
