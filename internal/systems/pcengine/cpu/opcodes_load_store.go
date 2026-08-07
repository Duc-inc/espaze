package cpu

func init() {
	setOp(0xA9, func(c *CPU) int { c.regs.A = c.readLoc(c.immediate()); c.regs.setNZ(c.regs.A); return 2 })
	setOp(0xA5, func(c *CPU) int { c.regs.A = c.readLoc(c.zeroPage()); c.regs.setNZ(c.regs.A); return 3 })
	setOp(0xB5, func(c *CPU) int { c.regs.A = c.readLoc(c.zeroPageX()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0xAD, func(c *CPU) int { c.regs.A = c.readLoc(c.absolute()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0xBD, func(c *CPU) int { c.regs.A = c.readLoc(c.absoluteX()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0xB9, func(c *CPU) int { c.regs.A = c.readLoc(c.absoluteY()); c.regs.setNZ(c.regs.A); return 4 })
	setOp(0xA1, func(c *CPU) int { c.regs.A = c.readLoc(c.indexedIndirect()); c.regs.setNZ(c.regs.A); return 6 })
	setOp(0xB1, func(c *CPU) int { c.regs.A = c.readLoc(c.indirectIndexed()); c.regs.setNZ(c.regs.A); return 5 })
	setOp(0xB2, func(c *CPU) int { c.regs.A = c.readLoc(c.zeroPageIndirect()); c.regs.setNZ(c.regs.A); return 5 })

	setOp(0xA2, func(c *CPU) int { c.regs.X = c.readLoc(c.immediate()); c.regs.setNZ(c.regs.X); return 2 })
	setOp(0xA6, func(c *CPU) int { c.regs.X = c.readLoc(c.zeroPage()); c.regs.setNZ(c.regs.X); return 3 })
	setOp(0xB6, func(c *CPU) int { c.regs.X = c.readLoc(c.zeroPageY()); c.regs.setNZ(c.regs.X); return 4 })
	setOp(0xAE, func(c *CPU) int { c.regs.X = c.readLoc(c.absolute()); c.regs.setNZ(c.regs.X); return 4 })
	setOp(0xBE, func(c *CPU) int { c.regs.X = c.readLoc(c.absoluteY()); c.regs.setNZ(c.regs.X); return 4 })

	setOp(0xA0, func(c *CPU) int { c.regs.Y = c.readLoc(c.immediate()); c.regs.setNZ(c.regs.Y); return 2 })
	setOp(0xA4, func(c *CPU) int { c.regs.Y = c.readLoc(c.zeroPage()); c.regs.setNZ(c.regs.Y); return 3 })
	setOp(0xB4, func(c *CPU) int { c.regs.Y = c.readLoc(c.zeroPageX()); c.regs.setNZ(c.regs.Y); return 4 })
	setOp(0xAC, func(c *CPU) int { c.regs.Y = c.readLoc(c.absolute()); c.regs.setNZ(c.regs.Y); return 4 })
	setOp(0xBC, func(c *CPU) int { c.regs.Y = c.readLoc(c.absoluteX()); c.regs.setNZ(c.regs.Y); return 4 })

	setOp(0x85, func(c *CPU) int { c.writeLoc(c.zeroPage(), c.regs.A); return 3 })
	setOp(0x95, func(c *CPU) int { c.writeLoc(c.zeroPageX(), c.regs.A); return 4 })
	setOp(0x8D, func(c *CPU) int { c.writeLoc(c.absolute(), c.regs.A); return 4 })
	setOp(0x9D, func(c *CPU) int { c.writeLoc(c.absoluteX(), c.regs.A); return 5 })
	setOp(0x99, func(c *CPU) int { c.writeLoc(c.absoluteY(), c.regs.A); return 5 })
	setOp(0x81, func(c *CPU) int { c.writeLoc(c.indexedIndirect(), c.regs.A); return 6 })
	setOp(0x91, func(c *CPU) int { c.writeLoc(c.indirectIndexed(), c.regs.A); return 6 })
	setOp(0x92, func(c *CPU) int { c.writeLoc(c.zeroPageIndirect(), c.regs.A); return 5 })

	setOp(0x86, func(c *CPU) int { c.writeLoc(c.zeroPage(), c.regs.X); return 3 })
	setOp(0x96, func(c *CPU) int { c.writeLoc(c.zeroPageY(), c.regs.X); return 4 })
	setOp(0x8E, func(c *CPU) int { c.writeLoc(c.absolute(), c.regs.X); return 4 })

	setOp(0x84, func(c *CPU) int { c.writeLoc(c.zeroPage(), c.regs.Y); return 3 })
	setOp(0x94, func(c *CPU) int { c.writeLoc(c.zeroPageX(), c.regs.Y); return 4 })
	setOp(0x8C, func(c *CPU) int { c.writeLoc(c.absolute(), c.regs.Y); return 4 })

	// STZ: the 65C02/HuC6280 addition that stores zero without needing
	// A to already hold it.
	setOp(0x64, func(c *CPU) int { c.writeLoc(c.zeroPage(), 0); return 3 })
	setOp(0x74, func(c *CPU) int { c.writeLoc(c.zeroPageX(), 0); return 4 })
	setOp(0x9C, func(c *CPU) int { c.writeLoc(c.absolute(), 0); return 4 })
	setOp(0x9E, func(c *CPU) int { c.writeLoc(c.absoluteX(), 0); return 5 })
}
