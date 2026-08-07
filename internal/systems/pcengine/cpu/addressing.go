package cpu

// location is an operand's resolved logical address, computed once so
// read-modify-write instructions don't re-evaluate (and potentially
// double-apply) an addressing mode's side effects.
type location struct {
	addr        uint16
	accumulator bool
}

func (c *CPU) readLoc(l location) byte {
	if l.accumulator {
		return c.regs.A
	}
	return c.read(l.addr)
}

func (c *CPU) writeLoc(l location, v byte) {
	if l.accumulator {
		c.regs.A = v
		return
	}
	c.write(l.addr, v)
}

func (c *CPU) zeroPage() location       { return location{addr: uint16(c.fetchByte())} }
func (c *CPU) zeroPageX() location      { return location{addr: uint16(c.fetchByte() + c.regs.X)} }
func (c *CPU) zeroPageY() location      { return location{addr: uint16(c.fetchByte() + c.regs.Y)} }
func (c *CPU) absolute() location       { return location{addr: c.fetch16()} }
func (c *CPU) absoluteX() location      { return location{addr: c.fetch16() + uint16(c.regs.X)} }
func (c *CPU) absoluteY() location      { return location{addr: c.fetch16() + uint16(c.regs.Y)} }
func (c *CPU) accumulatorLoc() location { return location{accumulator: true} }
func (c *CPU) immediate() location {
	l := location{addr: c.regs.PC}
	c.regs.PC++
	return l
}

// indexedIndirect implements (zp,X): the base+X page-0 byte pair holds
// the actual 16-bit address.
func (c *CPU) indexedIndirect() location {
	base := c.fetchByte() + c.regs.X
	lo := uint16(c.read(uint16(base)))
	hi := uint16(c.read(uint16(base + 1)))
	return location{addr: lo | hi<<8}
}

// indirectIndexed implements (zp),Y.
func (c *CPU) indirectIndexed() location {
	base := c.fetchByte()
	lo := uint16(c.read(uint16(base)))
	hi := uint16(c.read(uint16(base + 1)))
	return location{addr: (lo | hi<<8) + uint16(c.regs.Y)}
}

// zeroPageIndirect implements the 65C02/HuC6280 addition (zp), with
// no index register.
func (c *CPU) zeroPageIndirect() location {
	base := c.fetchByte()
	lo := uint16(c.read(uint16(base)))
	hi := uint16(c.read(uint16(base + 1)))
	return location{addr: lo | hi<<8}
}

func (c *CPU) relativeTarget() uint16 {
	offset := int8(c.fetchByte())
	return uint16(int32(c.regs.PC) + int32(offset))
}
