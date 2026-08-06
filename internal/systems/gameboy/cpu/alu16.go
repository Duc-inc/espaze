package cpu

// addHL implements ADD HL,rr: 16-bit add, Z is left untouched.
func (c *CPU) addHL(v uint16) {
	hl := c.regs.HL()
	sum := int(hl) + int(v)
	c.regs.SetFlag(FlagN, false)
	c.regs.SetFlag(FlagH, (hl&0xFFF)+(v&0xFFF) > 0xFFF)
	c.regs.SetFlag(FlagC, sum > 0xFFFF)
	c.regs.SetHL(uint16(sum))
}

// addSPSigned implements the shared arithmetic behind ADD SP,r8 (0xE8)
// and LD HL,SP+r8 (0xF8): both compute SP + a signed 8-bit offset, but
// set H/C from the *unsigned* low-byte addition per the LR35902 manual.
func (c *CPU) addSPSigned(offset int8) uint16 {
	sp := c.regs.SP
	result := uint16(int32(sp) + int32(offset))

	c.regs.SetFlag(FlagZ, false)
	c.regs.SetFlag(FlagN, false)
	c.regs.SetFlag(FlagH, (sp&0x000F)+uint16(byte(offset)&0x0F) > 0x000F)
	c.regs.SetFlag(FlagC, (sp&0x00FF)+uint16(byte(offset)) > 0x00FF)
	return result
}
