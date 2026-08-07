package cpu

// adc/sbc are always binary on the HuC6280 - a documented hardware
// quirk where, unlike the base 6502/65C02, the Decimal flag has no
// effect on these two instructions at all.
func adc(c *CPU, v byte) {
	sum := uint16(c.regs.A) + uint16(v)
	if c.regs.getFlag(FlagCarry) {
		sum++
	}
	result := byte(sum)
	c.regs.setFlag(FlagCarry, sum > 0xFF)
	c.regs.setFlag(FlagOverflow, (c.regs.A^v)&0x80 == 0 && (c.regs.A^result)&0x80 != 0)
	c.regs.A = result
	c.regs.setNZ(result)
}

func sbc(c *CPU, v byte) {
	borrow := byte(1)
	if c.regs.getFlag(FlagCarry) {
		borrow = 0
	}
	diff := int16(c.regs.A) - int16(v) - int16(borrow)
	result := byte(diff)
	c.regs.setFlag(FlagCarry, diff >= 0)
	c.regs.setFlag(FlagOverflow, (c.regs.A^v)&0x80 != 0 && (c.regs.A^result)&0x80 != 0)
	c.regs.A = result
	c.regs.setNZ(result)
}

func compare(c *CPU, reg byte, v byte) {
	result := reg - v
	c.regs.setFlag(FlagCarry, reg >= v)
	c.regs.setNZ(result)
}

func init() {
	setOp(0x69, func(c *CPU) int { adc(c, c.readLoc(c.immediate())); return 2 })
	setOp(0x65, func(c *CPU) int { adc(c, c.readLoc(c.zeroPage())); return 3 })
	setOp(0x75, func(c *CPU) int { adc(c, c.readLoc(c.zeroPageX())); return 4 })
	setOp(0x6D, func(c *CPU) int { adc(c, c.readLoc(c.absolute())); return 4 })
	setOp(0x7D, func(c *CPU) int { adc(c, c.readLoc(c.absoluteX())); return 4 })
	setOp(0x79, func(c *CPU) int { adc(c, c.readLoc(c.absoluteY())); return 4 })
	setOp(0x61, func(c *CPU) int { adc(c, c.readLoc(c.indexedIndirect())); return 6 })
	setOp(0x71, func(c *CPU) int { adc(c, c.readLoc(c.indirectIndexed())); return 5 })
	setOp(0x72, func(c *CPU) int { adc(c, c.readLoc(c.zeroPageIndirect())); return 5 })

	setOp(0xE9, func(c *CPU) int { sbc(c, c.readLoc(c.immediate())); return 2 })
	setOp(0xE5, func(c *CPU) int { sbc(c, c.readLoc(c.zeroPage())); return 3 })
	setOp(0xF5, func(c *CPU) int { sbc(c, c.readLoc(c.zeroPageX())); return 4 })
	setOp(0xED, func(c *CPU) int { sbc(c, c.readLoc(c.absolute())); return 4 })
	setOp(0xFD, func(c *CPU) int { sbc(c, c.readLoc(c.absoluteX())); return 4 })
	setOp(0xF9, func(c *CPU) int { sbc(c, c.readLoc(c.absoluteY())); return 4 })
	setOp(0xE1, func(c *CPU) int { sbc(c, c.readLoc(c.indexedIndirect())); return 6 })
	setOp(0xF1, func(c *CPU) int { sbc(c, c.readLoc(c.indirectIndexed())); return 5 })
	setOp(0xF2, func(c *CPU) int { sbc(c, c.readLoc(c.zeroPageIndirect())); return 5 })

	setOp(0xC9, func(c *CPU) int { compare(c, c.regs.A, c.readLoc(c.immediate())); return 2 })
	setOp(0xC5, func(c *CPU) int { compare(c, c.regs.A, c.readLoc(c.zeroPage())); return 3 })
	setOp(0xD5, func(c *CPU) int { compare(c, c.regs.A, c.readLoc(c.zeroPageX())); return 4 })
	setOp(0xCD, func(c *CPU) int { compare(c, c.regs.A, c.readLoc(c.absolute())); return 4 })
	setOp(0xDD, func(c *CPU) int { compare(c, c.regs.A, c.readLoc(c.absoluteX())); return 4 })
	setOp(0xD9, func(c *CPU) int { compare(c, c.regs.A, c.readLoc(c.absoluteY())); return 4 })
	setOp(0xC1, func(c *CPU) int { compare(c, c.regs.A, c.readLoc(c.indexedIndirect())); return 6 })
	setOp(0xD1, func(c *CPU) int { compare(c, c.regs.A, c.readLoc(c.indirectIndexed())); return 5 })

	setOp(0xE0, func(c *CPU) int { compare(c, c.regs.X, c.readLoc(c.immediate())); return 2 })
	setOp(0xE4, func(c *CPU) int { compare(c, c.regs.X, c.readLoc(c.zeroPage())); return 3 })
	setOp(0xEC, func(c *CPU) int { compare(c, c.regs.X, c.readLoc(c.absolute())); return 4 })

	setOp(0xC0, func(c *CPU) int { compare(c, c.regs.Y, c.readLoc(c.immediate())); return 2 })
	setOp(0xC4, func(c *CPU) int { compare(c, c.regs.Y, c.readLoc(c.zeroPage())); return 3 })
	setOp(0xCC, func(c *CPU) int { compare(c, c.regs.Y, c.readLoc(c.absolute())); return 4 })
}
