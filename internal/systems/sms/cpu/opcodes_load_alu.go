package cpu

// decodeX1 handles the x=1 opcode block: LD r,r' for every register
// pair (including memory via (HL)) except the one combination that
// means something else entirely - LD (HL),(HL) is encoded as HALT.
func (c *CPU) decodeX1(y, z byte) int {
	if y == 6 && z == 6 {
		c.halted = true
		return 4
	}
	c.setR8(y, c.r8(z))
	if y == 6 || z == 6 {
		return 7
	}
	return 4
}

// decodeX2 handles the x=2 opcode block: the 8 accumulator ALU
// operations against every register/memory operand.
func (c *CPU) decodeX2(y, z byte) int {
	c.aluOp(y, c.r8(z))
	if z == 6 {
		return 7
	}
	return 4
}
