package cpu

// subOpcodes registers SUB/SUBA/SUBI/SUBQ plus NEG and CLR, which share
// SUB's underlying subtract-with-flags logic.
func subOpcodes() []pattern {
	return []pattern{
		{mask: 0xF000, match: 0x9000, execute: subExecute}, // SUB / SUBA
		{mask: 0xFF00, match: 0x0400, execute: subiExecute},
		{mask: 0xF100, match: 0x5100, execute: subqExecute},
		{mask: 0xFF00, match: 0x4400, execute: negExecute},
		{mask: 0xFF00, match: 0x4200, execute: clrExecute},
	}
}

// subExecute mirrors addExecute exactly, just subtracting instead.
func subExecute(c *CPU, opcode uint16) int {
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
		c.regs.A[reg] -= src
		return 8
	}

	size := opmode & 0x03
	if opmode < 4 { // <ea> - Dn -> Dn... actually Dn - <ea> -> Dn per the SUB direction bit
		eaVal := c.readEA(mode, eaReg, size)
		dnVal := maskSize(c.regs.D[reg], size)
		result, carry, overflow := subWithFlags(dnVal, eaVal, size)
		c.regs.D[reg] = mergeSize(c.regs.D[reg], result, size)
		c.setArithFlags(result, size, carry, overflow)
		return 4
	}

	loc := c.resolveEA(mode, eaReg, size)
	eaVal := c.readLocation(loc, size)
	dnVal := maskSize(c.regs.D[reg], size)
	result, carry, overflow := subWithFlags(eaVal, dnVal, size)
	c.writeLocation(loc, size, result)
	c.setArithFlags(result, size, carry, overflow)
	return 8
}

func subiExecute(c *CPU, opcode uint16) int {
	size := byte((opcode >> 6) & 0x03)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	imm := c.fetchImmediate(size)
	loc := c.resolveEA(mode, eaReg, size)
	eaVal := c.readLocation(loc, size)

	result, carry, overflow := subWithFlags(eaVal, imm, size)
	c.writeLocation(loc, size, result)
	c.setArithFlags(result, size, carry, overflow)
	return 8
}

func subqExecute(c *CPU, opcode uint16) int {
	data := byte((opcode >> 9) & 0x07)
	if data == 0 {
		data = 8
	}
	size := byte((opcode >> 6) & 0x03)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	if mode == 1 {
		c.regs.A[eaReg] -= uint32(data)
		return 8
	}

	loc := c.resolveEA(mode, eaReg, size)
	eaVal := c.readLocation(loc, size)
	result, carry, overflow := subWithFlags(eaVal, uint32(data), size)
	c.writeLocation(loc, size, result)
	c.setArithFlags(result, size, carry, overflow)
	return 4
}

// negExecute implements NEG <ea>: 0 - <ea> -> <ea>.
func negExecute(c *CPU, opcode uint16) int {
	size := byte((opcode >> 6) & 0x03)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	loc := c.resolveEA(mode, eaReg, size)
	eaVal := c.readLocation(loc, size)
	result, carry, overflow := subWithFlags(0, eaVal, size)
	c.writeLocation(loc, size, result)
	c.setArithFlags(result, size, carry, overflow)
	return 4
}

// clrExecute implements CLR <ea>: zero it, Z set, everything else clear.
func clrExecute(c *CPU, opcode uint16) int {
	size := byte((opcode >> 6) & 0x03)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	loc := c.resolveEA(mode, eaReg, size)
	c.writeLocation(loc, size, 0)
	c.setLogicFlags(0, size)
	return 4
}
