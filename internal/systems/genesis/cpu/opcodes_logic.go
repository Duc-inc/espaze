package cpu

// logicOpcodes registers AND/ANDI, OR/ORI, EOR/EORI and NOT. AND and OR
// share their top nibble with MULU/MULS and DIVU/DIVS respectively
// (opmode 3 and 7) - those are claimed first by muldivOpcodes, which is
// registered earlier, so the broad patterns here only ever see the
// opmode values that are actually AND/OR. EOR needs the opposite
// treatment: it shares CMP/CMPA's top nibble, so *it* has to register
// narrow, exact patterns (opmode 4-6 specifically) ahead of
// compareOpcodes for the same reason.
func logicOpcodes() []pattern {
	return []pattern{
		{mask: 0xF000, match: 0xC000, execute: andExecute},
		{mask: 0xFF00, match: 0x0200, execute: andiExecute},
		{mask: 0xF000, match: 0x8000, execute: orExecute},
		{mask: 0xFF00, match: 0x0000, execute: oriExecute},
		{mask: 0xF1C0, match: 0xB100, execute: eorExecute}, // byte
		{mask: 0xF1C0, match: 0xB140, execute: eorExecute}, // word
		{mask: 0xF1C0, match: 0xB180, execute: eorExecute}, // long
		{mask: 0xFF00, match: 0x0A00, execute: eoriExecute},
		{mask: 0xFF00, match: 0x4600, execute: notExecute},
	}
}

// binaryLogicOp abstracts AND/OR's shared shape: opmode<4 reads <ea>,
// combines with Dn, writes back to Dn; opmode>=4 does the reverse.
func binaryLogicOp(c *CPU, opcode uint16, op func(a, b uint32) uint32) int {
	reg := byte((opcode >> 9) & 0x07)
	opmode := byte((opcode >> 6) & 0x07)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)
	size := opmode & 0x03

	if opmode < 4 {
		eaVal := c.readEA(mode, eaReg, size)
		dnVal := maskSize(c.regs.D[reg], size)
		result := op(dnVal, eaVal)
		c.regs.D[reg] = mergeSize(c.regs.D[reg], result, size)
		c.setLogicFlags(result, size)
		return 4
	}

	loc := c.resolveEA(mode, eaReg, size)
	eaVal := c.readLocation(loc, size)
	dnVal := maskSize(c.regs.D[reg], size)
	result := op(eaVal, dnVal)
	c.writeLocation(loc, size, result)
	c.setLogicFlags(result, size)
	return 8
}

func andExecute(c *CPU, opcode uint16) int {
	return binaryLogicOp(c, opcode, func(a, b uint32) uint32 { return a & b })
}

func orExecute(c *CPU, opcode uint16) int {
	return binaryLogicOp(c, opcode, func(a, b uint32) uint32 { return a | b })
}

func eorExecute(c *CPU, opcode uint16) int {
	reg := byte((opcode >> 9) & 0x07)
	opmode := byte((opcode >> 6) & 0x07)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)
	size := opmode & 0x03

	loc := c.resolveEA(mode, eaReg, size)
	eaVal := c.readLocation(loc, size)
	dnVal := maskSize(c.regs.D[reg], size)
	result := eaVal ^ dnVal
	c.writeLocation(loc, size, result)
	c.setLogicFlags(result, size)
	return 8
}

func immediateLogicOp(c *CPU, opcode uint16, op func(a, b uint32) uint32) int {
	size := byte((opcode >> 6) & 0x03)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	imm := c.fetchImmediate(size)
	loc := c.resolveEA(mode, eaReg, size)
	eaVal := c.readLocation(loc, size)
	result := op(eaVal, imm)
	c.writeLocation(loc, size, result)
	c.setLogicFlags(result, size)
	return 8
}

func andiExecute(c *CPU, opcode uint16) int {
	return immediateLogicOp(c, opcode, func(a, b uint32) uint32 { return a & b })
}

func oriExecute(c *CPU, opcode uint16) int {
	return immediateLogicOp(c, opcode, func(a, b uint32) uint32 { return a | b })
}

func eoriExecute(c *CPU, opcode uint16) int {
	return immediateLogicOp(c, opcode, func(a, b uint32) uint32 { return a ^ b })
}

func notExecute(c *CPU, opcode uint16) int {
	size := byte((opcode >> 6) & 0x03)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	loc := c.resolveEA(mode, eaReg, size)
	result := ^c.readLocation(loc, size)
	c.writeLocation(loc, size, result)
	c.setLogicFlags(result, size)
	return 4
}
