package cpu

// moveOpcodes covers MOVE and MOVEA - together the single largest slice
// of the 68000's opcode space (every source/destination addressing
// mode combination, in 3 sizes). Byte/word/long each get their own
// 4-bit-prefix pattern (0001/0011/0010) since the size field's bit
// pattern doesn't line up with this project's own byte/word/long
// constants.
func moveOpcodes() []pattern {
	return []pattern{
		{mask: 0xF000, match: 0x1000, execute: moveExecute}, // byte
		{mask: 0xF000, match: 0x3000, execute: moveExecute}, // word
		{mask: 0xF000, match: 0x2000, execute: moveExecute}, // long
		{mask: 0xF100, match: 0x7000, execute: moveqExecute},
		{mask: 0xF1C0, match: 0x41C0, execute: leaExecute},
		{mask: 0xFFC0, match: 0x4840, execute: peaExecute},
	}
}

func moveSizeFromOpcode(opcode uint16) byte {
	switch (opcode >> 12) & 0x03 {
	case 1:
		return sizeByte
	case 3:
		return sizeWord
	default:
		return sizeLong
	}
}

func moveExecute(c *CPU, opcode uint16) int {
	size := moveSizeFromOpcode(opcode)
	srcReg := byte(opcode & 0x07)
	srcMode := byte((opcode >> 3) & 0x07)
	destReg := byte((opcode >> 9) & 0x07)
	destMode := byte((opcode >> 6) & 0x07)

	value := c.readEA(srcMode, srcReg, size)
	destLoc := c.resolveEA(destMode, destReg, size)
	c.writeLocation(destLoc, size, value)

	if destMode != 1 { // MOVEA (destMode 1) never touches flags
		c.regs.setFlag(FlagN, isNegativeSized(value, size))
		c.regs.setFlag(FlagZ, maskSize(value, size) == 0)
		c.regs.setFlag(FlagV, false)
		c.regs.setFlag(FlagC, false)
	}
	return 4
}

func isNegativeSized(v uint32, size byte) bool {
	switch size {
	case sizeByte:
		return int8(v) < 0
	case sizeLong:
		return int32(v) < 0
	default:
		return int16(v) < 0
	}
}

// moveqExecute implements MOVEQ #imm,Dn: an 8-bit immediate sign-extended
// into a data register, one of the cheapest ways to load a small
// constant.
func moveqExecute(c *CPU, opcode uint16) int {
	reg := byte((opcode >> 9) & 0x07)
	imm := int32(int8(opcode))
	c.regs.D[reg] = uint32(imm)
	c.regs.setFlag(FlagN, imm < 0)
	c.regs.setFlag(FlagZ, imm == 0)
	c.regs.setFlag(FlagV, false)
	c.regs.setFlag(FlagC, false)
	return 4
}

// leaExecute implements LEA <ea>,An: loads an effective *address*
// (never dereferencing it) into an address register.
func leaExecute(c *CPU, opcode uint16) int {
	mode := byte((opcode >> 3) & 0x07)
	reg := byte(opcode & 0x07)
	destReg := byte((opcode >> 9) & 0x07)

	loc := c.resolveEA(mode, reg, sizeLong)
	c.regs.A[destReg] = loc.addr
	return 4
}

// peaExecute implements PEA <ea>: pushes an effective address onto the
// stack, without touching any register.
func peaExecute(c *CPU, opcode uint16) int {
	mode := byte((opcode >> 3) & 0x07)
	reg := byte(opcode & 0x07)
	loc := c.resolveEA(mode, reg, sizeLong)
	c.push32(loc.addr)
	return 12
}
