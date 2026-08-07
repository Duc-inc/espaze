package cpu

// arithmeticOpcodes registers ADD/ADDA/ADDI/ADDQ and their SUB
// counterparts (see opcodes_sub.go for the latter's patterns, appended
// by the same slice this function returns).
func arithmeticOpcodes() []pattern {
	p := []pattern{
		{mask: 0xF000, match: 0xD000, execute: addExecute}, // ADD / ADDA
		{mask: 0xFF00, match: 0x0600, execute: addiExecute},
		{mask: 0xF100, match: 0x5000, execute: addqExecute}, // bit8=0 distinguishes ADDQ from SUBQ
	}
	return append(p, subOpcodes()...)
}

// addExecute implements ADD (opmodes 0-2 and 4-6) and ADDA (opmodes 3
// and 7), which share the same top nibble.
func addExecute(c *CPU, opcode uint16) int {
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
		c.regs.A[reg] += src
		return 8
	}

	size := opmode & 0x03
	if opmode < 4 { // <ea> + Dn -> Dn
		eaVal := c.readEA(mode, eaReg, size)
		dnVal := maskSize(c.regs.D[reg], size)
		result, carry, overflow := addWithFlags(dnVal, eaVal, size)
		c.regs.D[reg] = mergeSize(c.regs.D[reg], result, size)
		c.setArithFlags(result, size, carry, overflow)
		return 4
	}

	// Dn + <ea> -> <ea>
	loc := c.resolveEA(mode, eaReg, size)
	eaVal := c.readLocation(loc, size)
	dnVal := maskSize(c.regs.D[reg], size)
	result, carry, overflow := addWithFlags(eaVal, dnVal, size)
	c.writeLocation(loc, size, result)
	c.setArithFlags(result, size, carry, overflow)
	return 8
}

// addiExecute implements ADDI #imm,<ea>: size comes directly from bits
// 7-6, which luckily already matches this project's size constants.
func addiExecute(c *CPU, opcode uint16) int {
	size := byte((opcode >> 6) & 0x03)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	imm := c.fetchImmediate(size)
	loc := c.resolveEA(mode, eaReg, size)
	eaVal := c.readLocation(loc, size)

	result, carry, overflow := addWithFlags(eaVal, imm, size)
	c.writeLocation(loc, size, result)
	c.setArithFlags(result, size, carry, overflow)
	return 8
}

// addqExecute implements ADDQ #data,<ea>: a 3-bit immediate (0 means 8)
// added directly, with no extension word - one of the cheapest ways to
// bump a counter or pointer. Targeting an address register is a special
// case: always a full 32-bit add, and it never touches the flags.
func addqExecute(c *CPU, opcode uint16) int {
	data := byte((opcode >> 9) & 0x07)
	if data == 0 {
		data = 8
	}
	size := byte((opcode >> 6) & 0x03)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)

	if mode == 1 {
		c.regs.A[eaReg] += uint32(data)
		return 8
	}

	loc := c.resolveEA(mode, eaReg, size)
	eaVal := c.readLocation(loc, size)
	result, carry, overflow := addWithFlags(eaVal, uint32(data), size)
	c.writeLocation(loc, size, result)
	c.setArithFlags(result, size, carry, overflow)
	return 4
}
