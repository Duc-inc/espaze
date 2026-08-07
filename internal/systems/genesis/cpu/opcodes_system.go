package cpu

// systemOpcodes registers everything that has to win precedence over a
// broader pattern elsewhere (SWAP/EXT over PEA/MOVEM, EXG over AND, the
// CCR/SR immediate forms over plain ANDI/ORI/EORI - see each pattern's
// neighbor for why), plus NOP/RESET/STOP and USP access.
func systemOpcodes() []pattern {
	return []pattern{
		{mask: 0xFFFF, match: 0x003C, execute: oriCCRExecute},
		{mask: 0xFFFF, match: 0x007C, execute: oriSRExecute},
		{mask: 0xFFFF, match: 0x023C, execute: andiCCRExecute},
		{mask: 0xFFFF, match: 0x027C, execute: andiSRExecute},
		{mask: 0xFFFF, match: 0x0A3C, execute: eoriCCRExecute},
		{mask: 0xFFFF, match: 0x0A7C, execute: eoriSRExecute},
		{mask: 0xFFC0, match: 0x40C0, execute: moveFromSRExecute},
		{mask: 0xFFC0, match: 0x44C0, execute: moveToCCRExecute},
		{mask: 0xFFC0, match: 0x46C0, execute: moveToSRExecute},
		{mask: 0xFFF8, match: 0x4840, execute: swapExecute},
		{mask: 0xFFF8, match: 0x4880, execute: extExecute},
		{mask: 0xFFF8, match: 0x48C0, execute: extExecute},
		{mask: 0xF1F8, match: 0xC140, execute: exgExecute},
		{mask: 0xF1F8, match: 0xC148, execute: exgExecute},
		{mask: 0xF1F8, match: 0xC188, execute: exgExecute},
		{mask: 0xFFF0, match: 0x4E60, execute: moveUSPExecute},
		{mask: 0xFFFF, match: 0x4E71, execute: nopExecute},
		{mask: 0xFFFF, match: 0x4E70, execute: nopExecute}, // RESET: no external hardware to reset here
		{mask: 0xFFFF, match: 0x4E76, execute: nopExecute}, // TRAPV: no V-overflow trap tracking
		{mask: 0xFFFF, match: 0x4E72, execute: stopExecute},
	}
}

func withCCR(c *CPU, op func(sr, v uint16) uint16) int {
	imm := uint16(c.fetchWord()) & 0x00FF
	c.regs.SR = op(c.regs.SR, imm)&0x00FF | c.regs.SR&0xFF00
	return 20
}

func oriCCRExecute(c *CPU, opcode uint16) int {
	return withCCR(c, func(sr, v uint16) uint16 { return sr | v })
}
func andiCCRExecute(c *CPU, opcode uint16) int {
	return withCCR(c, func(sr, v uint16) uint16 { return sr & v })
}
func eoriCCRExecute(c *CPU, opcode uint16) int {
	return withCCR(c, func(sr, v uint16) uint16 { return sr ^ v })
}

func withSR(c *CPU, op func(sr, v uint16) uint16) int {
	imm := c.fetchWord()
	c.regs.SR = op(c.regs.SR, imm)
	return 20
}

func oriSRExecute(c *CPU, opcode uint16) int {
	return withSR(c, func(sr, v uint16) uint16 { return sr | v })
}
func andiSRExecute(c *CPU, opcode uint16) int {
	return withSR(c, func(sr, v uint16) uint16 { return sr & v })
}
func eoriSRExecute(c *CPU, opcode uint16) int {
	return withSR(c, func(sr, v uint16) uint16 { return sr ^ v })
}

func moveFromSRExecute(c *CPU, opcode uint16) int {
	mode := byte((opcode >> 3) & 0x07)
	reg := byte(opcode & 0x07)
	loc := c.resolveEA(mode, reg, sizeWord)
	c.writeLocation(loc, sizeWord, uint32(c.regs.SR))
	return 8
}

func moveToCCRExecute(c *CPU, opcode uint16) int {
	mode := byte((opcode >> 3) & 0x07)
	reg := byte(opcode & 0x07)
	v := c.readEA(mode, reg, sizeWord)
	c.regs.SR = c.regs.SR&0xFF00 | uint16(v)&0x00FF
	return 12
}

func moveToSRExecute(c *CPU, opcode uint16) int {
	mode := byte((opcode >> 3) & 0x07)
	reg := byte(opcode & 0x07)
	v := c.readEA(mode, reg, sizeWord)
	c.regs.SR = uint16(v)
	return 12
}

// swapExecute implements SWAP Dn: exchanges the high and low 16-bit
// halves of a data register.
func swapExecute(c *CPU, opcode uint16) int {
	reg := byte(opcode & 0x07)
	v := c.regs.D[reg]
	c.regs.D[reg] = v<<16 | v>>16
	c.regs.setFlag(FlagN, int32(c.regs.D[reg]) < 0)
	c.regs.setFlag(FlagZ, c.regs.D[reg] == 0)
	c.regs.setFlag(FlagV, false)
	c.regs.setFlag(FlagC, false)
	return 4
}

// extExecute implements EXT: sign-extends a data register's low byte
// to a word, or its low word to a long, depending on opcode bit 6.
func extExecute(c *CPU, opcode uint16) int {
	reg := byte(opcode & 0x07)
	toLong := opcode&0x0040 != 0
	if toLong {
		c.regs.D[reg] = uint32(int32(int16(c.regs.D[reg])))
		c.setLogicFlags(c.regs.D[reg], sizeLong)
	} else {
		v := mergeSize(c.regs.D[reg], uint32(uint16(int16(int8(c.regs.D[reg])))), sizeWord)
		c.regs.D[reg] = v
		c.setLogicFlags(v, sizeWord)
	}
	return 4
}

// exgExecute implements EXG: swaps the full 32-bit contents of two
// registers, in any Dx/Dy, Ax/Ay or Dx/Ay combination.
func exgExecute(c *CPU, opcode uint16) int {
	rx := byte((opcode >> 9) & 0x07)
	ry := byte(opcode & 0x07)
	mode := byte((opcode >> 3) & 0x1F)

	switch mode {
	case 0x08: // Dx,Dy
		c.regs.D[rx], c.regs.D[ry] = c.regs.D[ry], c.regs.D[rx]
	case 0x09: // Ax,Ay
		c.regs.A[rx], c.regs.A[ry] = c.regs.A[ry], c.regs.A[rx]
	default: // Dx,Ay
		c.regs.D[rx], c.regs.A[ry] = c.regs.A[ry], c.regs.D[rx]
	}
	return 6
}

// moveUSPExecute implements MOVE An,USP and MOVE USP,An - only ever
// meaningful in supervisor mode, which every ROM this project runs
// already is by the time it touches USP.
func moveUSPExecute(c *CPU, opcode uint16) int {
	reg := byte(opcode & 0x07)
	toUSP := opcode&0x0008 == 0
	if toUSP {
		c.regs.usp = c.regs.A[reg]
	} else {
		c.regs.A[reg] = c.regs.usp
	}
	return 4
}

func nopExecute(c *CPU, opcode uint16) int { return 4 }

// stopExecute implements STOP #imm: loads SR and halts the CPU until an
// interrupt (or reset) wakes it.
func stopExecute(c *CPU, opcode uint16) int {
	c.regs.SR = c.fetchWord()
	c.stopped = true
	return 4
}
