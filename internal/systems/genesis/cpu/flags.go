package cpu

func signBitFor(size byte) uint32 {
	switch size {
	case sizeByte:
		return 0x80
	case sizeLong:
		return 0x80000000
	default:
		return 0x8000
	}
}

// addWithFlags computes a+b at the given size and reports the carry-out
// and signed overflow real hardware would - both operands are masked
// to size first, so callers don't need to pre-truncate.
func addWithFlags(a, b uint32, size byte) (result uint32, carry, overflow bool) {
	a, b = maskSize(a, size), maskSize(b, size)
	sum := uint64(a) + uint64(b)
	bit := signBitFor(size)

	result = maskSize(uint32(sum), size)
	carry = sum > uint64(maskSize(0xFFFFFFFF, size))
	overflow = (a^b)&bit == 0 && (a^result)&bit != 0
	return
}

// subWithFlags computes a-b, matching CMP/SUB's convention.
func subWithFlags(a, b uint32, size byte) (result uint32, carry, overflow bool) {
	a, b = maskSize(a, size), maskSize(b, size)
	diff := int64(a) - int64(b)
	bit := signBitFor(size)

	result = maskSize(uint32(diff), size)
	carry = diff < 0
	overflow = (a^b)&(a^result)&bit != 0
	return
}

// setArithFlags applies the flag pattern ADD/SUB (and their immediate/
// quick variants) share: N/Z from the result, V/C from the operation,
// and X mirroring C (68000 arithmetic keeps Extend in lockstep with
// Carry, unlike logic ops which leave X alone).
func (c *CPU) setArithFlags(result uint32, size byte, carry, overflow bool) {
	c.regs.setFlag(FlagN, isNegativeSized(result, size))
	c.regs.setFlag(FlagZ, maskSize(result, size) == 0)
	c.regs.setFlag(FlagV, overflow)
	c.regs.setFlag(FlagC, carry)
	c.regs.setFlag(FlagX, carry)
}

// setLogicFlags applies AND/OR/EOR/NOT/MOVE's simpler pattern: N/Z from
// the result, V and C always cleared, X left untouched.
func (c *CPU) setLogicFlags(result uint32, size byte) {
	c.regs.setFlag(FlagN, isNegativeSized(result, size))
	c.regs.setFlag(FlagZ, maskSize(result, size) == 0)
	c.regs.setFlag(FlagV, false)
	c.regs.setFlag(FlagC, false)
}
