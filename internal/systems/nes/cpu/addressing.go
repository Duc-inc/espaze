package cpu

// addrMode identifies one of the 6502's addressing modes - how an
// instruction's operand bytes turn into the effective address (or
// non-address, for modeImplied/modeAccumulator) it operates on.
type addrMode int

const (
	modeImplied addrMode = iota
	modeAccumulator
	modeImmediate
	modeZeroPage
	modeZeroPageX
	modeZeroPageY
	modeAbsolute
	modeAbsoluteX
	modeAbsoluteY
	modeIndirect
	modeIndirectX
	modeIndirectY
	modeRelative
)

// resolveOperand consumes this instruction's operand bytes from the
// program counter and returns the effective address (0 for modes that
// have none) plus whether a page boundary was crossed while computing
// it - some instructions charge an extra cycle for that.
func (c *CPU) resolveOperand(mode addrMode) (uint16, bool) {
	switch mode {
	case modeImplied, modeAccumulator:
		return 0, false

	case modeImmediate:
		addr := c.regs.PC
		c.regs.PC++
		return addr, false

	case modeZeroPage:
		return uint16(c.fetchByte()), false

	case modeZeroPageX:
		return uint16(c.fetchByte() + c.regs.X), false

	case modeZeroPageY:
		return uint16(c.fetchByte() + c.regs.Y), false

	case modeAbsolute:
		return c.fetch16(), false

	case modeAbsoluteX:
		base := c.fetch16()
		addr := base + uint16(c.regs.X)
		return addr, pageDiffers(base, addr)

	case modeAbsoluteY:
		base := c.fetch16()
		addr := base + uint16(c.regs.Y)
		return addr, pageDiffers(base, addr)

	case modeIndirect:
		// Real 6502 hardware bug, preserved deliberately: if the pointer
		// sits at a page boundary (e.g. $xxFF), the high byte wraps
		// within that same page instead of crossing into the next one.
		ptr := c.fetch16()
		return c.read16(ptr), false

	case modeIndirectX:
		base := c.fetchByte() + c.regs.X
		lo := uint16(c.bus.Read(uint16(base)))
		hi := uint16(c.bus.Read(uint16(base + 1)))
		return hi<<8 | lo, false

	case modeIndirectY:
		zp := c.fetchByte()
		lo := uint16(c.bus.Read(uint16(zp)))
		hi := uint16(c.bus.Read(uint16(zp + 1)))
		base := hi<<8 | lo
		addr := base + uint16(c.regs.Y)
		return addr, pageDiffers(base, addr)

	case modeRelative:
		offset := int8(c.fetchByte())
		addr := uint16(int32(c.regs.PC) + int32(offset))
		return addr, pageDiffers(c.regs.PC, addr)
	}
	return 0, false
}

func pageDiffers(a, b uint16) bool {
	return a&0xFF00 != b&0xFF00
}
