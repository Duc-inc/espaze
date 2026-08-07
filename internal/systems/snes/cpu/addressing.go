package cpu

// location is an operand's resolved 24-bit effective address, computed
// once so read-modify-write instructions don't re-evaluate (and
// potentially double-apply) an addressing mode's side effects. An
// immediate operand (which has no memory address at all) is
// represented as its own variant so every ALU instruction's width-
// aware handler can consume it through the same readLoc8/16 calls as
// every other addressing mode, rather than needing a separate
// immediate-only code path per instruction.
type location struct {
	addr        uint32
	accumulator bool
	immediate   bool
	value       uint16
}

func (c *CPU) immediateLoc(v uint16) location { return location{immediate: true, value: v} }

func (c *CPU) readLoc8(l location) byte {
	switch {
	case l.immediate:
		return byte(l.value)
	case l.accumulator:
		return byte(c.regs.A)
	default:
		return c.read8(l.addr)
	}
}

func (c *CPU) readLoc16(l location) uint16 {
	switch {
	case l.immediate:
		return l.value
	case l.accumulator:
		return c.regs.A
	default:
		return c.read16(l.addr)
	}
}

func (c *CPU) writeLoc8(l location, v byte) {
	if l.accumulator {
		c.regs.A = c.regs.A&0xFF00 | uint16(v)
		return
	}
	c.write8(l.addr, v)
}

func (c *CPU) writeLoc16(l location, v uint16) {
	if l.accumulator {
		c.regs.A = v
		return
	}
	c.write16(l.addr, v)
}

func (c *CPU) accumulatorLoc() location { return location{accumulator: true} }

// Direct Page addressing always operates within bank 0, wrapping at
// the 16-bit boundary - real hardware's own behavior.
func (c *CPU) directPage() location {
	return location{addr: uint32(c.regs.D + uint16(c.fetch8()))}
}

func (c *CPU) directPageX() location {
	return location{addr: uint32(c.regs.D + uint16(c.fetch8()) + c.regs.X)}
}

func (c *CPU) directPageY() location {
	return location{addr: uint32(c.regs.D + uint16(c.fetch8()) + c.regs.Y)}
}

// directPageIndirect implements (dp): the 16-bit pointer stored at
// D+operand (bank 0) is combined with DBR for the actual bank.
func (c *CPU) directPageIndirect() location {
	ptrAddr := uint32(c.regs.D + uint16(c.fetch8()))
	ptr := c.read16(ptrAddr)
	return location{addr: uint32(c.regs.DBR)<<16 | uint32(ptr)}
}

// directPageIndirectY implements (dp),Y.
func (c *CPU) directPageIndirectY() location {
	ptrAddr := uint32(c.regs.D + uint16(c.fetch8()))
	ptr := c.read16(ptrAddr) + c.regs.Y
	return location{addr: uint32(c.regs.DBR)<<16 | uint32(ptr)}
}

// directPageIndirectLong implements [dp]: a 24-bit pointer, own bank included.
func (c *CPU) directPageIndirectLong() location {
	ptrAddr := uint32(c.regs.D + uint16(c.fetch8()))
	lo := uint32(c.read16(ptrAddr))
	hi := uint32(c.read8(ptrAddr + 2))
	return location{addr: lo | hi<<16}
}

// directPageIndirectLongY implements [dp],Y - unlike the short form,
// this add can carry into the bank byte.
func (c *CPU) directPageIndirectLongY() location {
	l := c.directPageIndirectLong()
	return location{addr: (l.addr + uint32(c.regs.Y)) & 0xFFFFFF}
}

func (c *CPU) absolute() location {
	return location{addr: uint32(c.regs.DBR)<<16 | uint32(c.fetch16())}
}

func (c *CPU) absoluteX() location {
	return location{addr: uint32(c.regs.DBR)<<16 | uint32(c.fetch16()+c.regs.X)}
}

func (c *CPU) absoluteY() location {
	return location{addr: uint32(c.regs.DBR)<<16 | uint32(c.fetch16()+c.regs.Y)}
}

func (c *CPU) absoluteLong() location {
	return location{addr: c.fetch24()}
}

func (c *CPU) absoluteLongX() location {
	return location{addr: (c.fetch24() + uint32(c.regs.X)) & 0xFFFFFF}
}

func (c *CPU) relativeTarget8() uint16 {
	offset := int8(c.fetch8())
	return uint16(int32(c.regs.PC) + int32(offset))
}

func (c *CPU) relativeTarget16() uint16 {
	offset := int16(c.fetch16())
	return uint16(int32(c.regs.PC) + int32(offset))
}
