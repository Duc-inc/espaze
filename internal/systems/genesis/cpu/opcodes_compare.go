package cpu

// compareOpcodes registers CMP/CMPA/CMPI and TST - all flags-only, none
// ever write their "result" anywhere.
func compareOpcodes() []pattern {
	return []pattern{
		{mask: 0xF000, match: 0xB000, execute: cmpExecute}, // CMP / CMPA
		{mask: 0xFF00, match: 0x0C00, execute: cmpiExecute},
		{mask: 0xFF00, match: 0x4A00, execute: tstExecute},
	}
}

func cmpExecute(c *CPU, opcode uint16) int {
	reg := byte((opcode >> 9) & 0x07)
	opmode := byte((opcode >> 6) & 0x07)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	if opmode == 3 || opmode == 7 {
		size := sizeWord
		if opmode == 7 {
			size = sizeLong
		}
		src := signExtend(c.readEA(mode, eaReg, size), size)
		result, carry, overflow := subWithFlags(c.regs.A[reg], src, sizeLong)
		c.setArithFlags(result, sizeLong, carry, overflow)
		c.regs.setFlag(FlagX, c.regs.getFlag(FlagX)) // CMPA doesn't touch X - restore it
		return 6
	}

	size := opmode & 0x03
	eaVal := c.readEA(mode, eaReg, size)
	dnVal := maskSize(c.regs.D[reg], size)
	xBefore := c.regs.getFlag(FlagX)
	result, carry, overflow := subWithFlags(dnVal, eaVal, size)
	c.setArithFlags(result, size, carry, overflow)
	c.regs.setFlag(FlagX, xBefore) // CMP doesn't touch X either
	return 4
}

func cmpiExecute(c *CPU, opcode uint16) int {
	size := byte((opcode >> 6) & 0x03)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	imm := c.fetchImmediate(size)
	eaVal := c.readEA(mode, eaReg, size)
	xBefore := c.regs.getFlag(FlagX)
	result, carry, overflow := subWithFlags(eaVal, imm, size)
	c.setArithFlags(result, size, carry, overflow)
	c.regs.setFlag(FlagX, xBefore)
	return 8
}

// tstExecute implements TST <ea>: like CMP against zero, but doesn't
// even touch V/C/X beyond what comparing-to-zero naturally implies.
func tstExecute(c *CPU, opcode uint16) int {
	size := byte((opcode >> 6) & 0x03)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	v := c.readEA(mode, eaReg, size)
	c.setLogicFlags(v, size)
	return 4
}
