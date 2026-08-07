package cpu

// thumbHiReg implements format 5: ADD/CMP/MOV on any register pair
// (including R8-R15, via the H1/H2 extension bits) and BX/BLX.
func (c *CPU) thumbHiReg(op uint16) int {
	rd := uint32(op&0x07) | uint32(op>>4)&0x08
	rs := uint32((op>>3)&0x07) | uint32(op>>6)&0x08

	switch (op >> 8) & 0x03 {
	case 0: // ADD
		c.regs.R[rd] = c.doAdd(c.regs.R[rd], c.regs.R[rs], false)
	case 1: // CMP
		c.doSub(c.regs.R[rd], c.regs.R[rs], true)
	case 2: // MOV
		c.regs.R[rd] = c.regs.R[rs]
	default: // BX/BLX
		target := c.regs.R[rs]
		c.regs.setFlag(FlagThumb, target&1 != 0)
		c.regs.R[15] = target &^ 1
	}
	if rd == 15 && (op>>8)&0x03 != 3 {
		c.regs.R[15] &^= 1
	}
	return 2
}

// thumbPCRelativeLoad implements format 6: LDR Rd,[PC,#imm8*4].
func (c *CPU) thumbPCRelativeLoad(op uint16) int {
	rd := (op >> 8) & 0x07
	imm := uint32(op&0xFF) * 4
	base := (c.regs.R[15] + 2) &^ 3
	c.regs.R[rd] = c.read32(base + imm)
	return 3
}

// thumbCondBranch implements format 16.
func (c *CPU) thumbCondBranch(op uint16) int {
	cond := uint32((op >> 8) & 0x0F)
	if !c.checkCondition(cond) {
		return 1
	}
	offset := int32(int8(byte(op)))
	c.regs.R[15] = uint32(int32(c.regs.R[15]+2) + offset*2)
	return 3
}

// thumbSWI implements format 17. This project has no BIOS to
// redistribute, so games relying on BIOS-provided SWI services
// (memory copy/fill, math routines, decompression) won't function
// correctly through this path - documented, not silently papered over.
func (c *CPU) thumbSWI(op uint16) int {
	return c.enterException(modeSupervisor, 0x08, 2)
}

// thumbBranch implements format 18.
func (c *CPU) thumbBranch(op uint16) int {
	offset := signExtend(uint32(op&0x7FF), 11)
	c.regs.R[15] = uint32(int32(c.regs.R[15]+2) + offset*2)
	return 3
}

// thumbBranchLink implements format 19: a two-halfword BL, executed
// across two Step calls - the first latches the upper 11 bits, the
// second combines them with its own lower 11 bits and jumps.
func (c *CPU) thumbBranchLink(op uint16) int {
	if op&0x0800 == 0 { // first halfword
		offsetHigh := signExtend(uint32(op&0x7FF), 11)
		c.regs.R[14] = uint32(int32(c.regs.R[15]+2) + offsetHigh*4096)
		return 2
	}
	// second halfword
	offsetLow := uint32(op&0x7FF) << 1
	target := c.regs.R[14] + offsetLow
	c.regs.R[14] = (c.regs.R[15] - 2 + 2) | 1
	c.regs.R[15] = target
	return 3
}

func signExtend(v uint32, bits uint) int32 {
	shift := 32 - bits
	return int32(v<<shift) >> shift
}
