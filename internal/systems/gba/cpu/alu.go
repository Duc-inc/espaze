package cpu

// addFlags/subFlags compute a 32-bit add/subtract along with the
// carry and overflow flags real hardware derives from it - shared by
// both the ARM and Thumb decoders' arithmetic instructions.
func addFlags(a, b uint32) (result uint32, carry, overflow bool) {
	sum := uint64(a) + uint64(b)
	result = uint32(sum)
	carry = sum > 0xFFFFFFFF
	overflow = (a^result)&(b^result)&0x80000000 != 0
	return
}

func subFlags(a, b uint32) (result uint32, carry, overflow bool) {
	result, carry, _ = addFlags(a, ^b+1)
	carry = a >= b
	overflow = (a^b)&(a^result)&0x80000000 != 0
	return
}

func (c *CPU) doAdd(a, b uint32, updateFlags bool) uint32 {
	result, carry, overflow := addFlags(a, b)
	if updateFlags {
		c.setNZ(result)
		c.regs.setFlag(FlagC, carry)
		c.regs.setFlag(FlagV, overflow)
	}
	return result
}

func (c *CPU) doSub(a, b uint32, updateFlags bool) uint32 {
	result, carry, overflow := subFlags(a, b)
	if updateFlags {
		c.setNZ(result)
		c.regs.setFlag(FlagC, carry)
		c.regs.setFlag(FlagV, overflow)
	}
	return result
}
