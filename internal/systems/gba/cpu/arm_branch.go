package cpu

// armBX implements Branch and Exchange: jump to Rm, switching to
// Thumb state if its low bit is set.
func (c *CPU) armBX(op uint32) int {
	target := c.regs.R[op&0x0F]
	c.regs.setFlag(FlagThumb, target&1 != 0)
	c.regs.R[15] = target &^ 1
	return 3
}

// armBranch implements B/BL: a signed 24-bit word offset relative to
// PC+8 (the real hardware's 2-stage-ahead prefetch value; this
// project's PC has already advanced by 4 from fetch32, so +4 more
// reproduces the same effective base).
func (c *CPU) armBranch(op uint32) int {
	offset := signExtend(op&0x00FFFFFF, 24) * 4
	if op&0x01000000 != 0 { // link bit
		c.regs.R[14] = c.regs.R[15]
	}
	c.regs.R[15] = uint32(int32(c.regs.R[15]+4) + offset)
	return 3
}

// armMultiply implements MUL/MLA (32-bit result only - the long
// multiply variants producing a 64-bit result aren't implemented).
func (c *CPU) armMultiply(op uint32) int {
	rd := (op >> 16) & 0x0F
	rn := (op >> 12) & 0x0F
	rs := (op >> 8) & 0x0F
	rm := op & 0x0F
	accumulate := op&0x00200000 != 0
	setFlags := op&0x00100000 != 0

	result := c.regs.R[rm] * c.regs.R[rs]
	if accumulate {
		result += c.regs.R[rn]
	}
	c.regs.R[rd] = result
	if setFlags {
		c.setNZ(result)
	}
	return 2
}
