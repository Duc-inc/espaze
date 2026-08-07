package cpu

func init() {
	setOp(0x00, func(c *CPU) int { return 2 }) // NOP
	setOp(0x01, func(c *CPU) int { c.halted = true; return 2 })
	setOp(0x02, func(c *CPU) int { c.regs.SR &^= 0x7000; return 2 }) // EI
	setOp(0x03, func(c *CPU) int { c.regs.SR |= 0x7000; return 2 })  // DI

	// LD r8,#imm8
	for r := byte(0); r < 8; r++ {
		reg := r
		setOp(0x10+reg, func(c *CPU) int {
			c.regs.setReg8(reg, c.fetch8())
			return 3
		})
	}

	// LD r8,(HL) / LD (HL),r8
	for r := byte(0); r < 8; r++ {
		reg := r
		setOp(0x20+reg, func(c *CPU) int {
			c.regs.setReg8(reg, c.bus.Read8(c.hlAddr()))
			return 4
		})
		setOp(0x28+reg, func(c *CPU) int {
			c.bus.Write8(c.hlAddr(), c.regs.reg8(reg))
			return 4
		})
	}

	// LD rp,#imm16 (WA/BC/DE/HL)
	for p := byte(0); p < 4; p++ {
		pair := p
		setOp(0x30+pair, func(c *CPU) int {
			c.regs.setReg16(pair, c.fetch16())
			return 4
		})
	}
	setOp(0x38, func(c *CPU) int { c.regs.XIX = uint32(c.fetch16()); return 4 })
	setOp(0x39, func(c *CPU) int { c.regs.XIY = uint32(c.fetch16()); return 4 })
	setOp(0x3A, func(c *CPU) int { c.regs.XIZ = uint32(c.fetch16()); return 4 })
	setOp(0x3B, func(c *CPU) int { c.regs.XSP = uint32(c.fetch16()); return 4 })

	// LD rp,(HL) / LD (HL),rp
	for p := byte(0); p < 4; p++ {
		pair := p
		setOp(0x3C+pair, func(c *CPU) int {
			c.regs.setReg16(pair, c.bus.Read16(c.hlAddr()))
			return 5
		})
		setOp(0x40+pair, func(c *CPU) int {
			c.bus.Write16(c.hlAddr(), c.regs.reg16(pair))
			return 5
		})
	}

	// LD r8dst,r8src and LD rp dst,rp src (operand byte: high nibble
	// dest, low nibble source).
	setOp(0x48, func(c *CPU) int {
		operand := c.fetch8()
		c.regs.setReg8(operand>>4, c.regs.reg8(operand&0x0F))
		return 3
	})
	setOp(0x49, func(c *CPU) int {
		operand := c.fetch8()
		c.regs.setReg16(operand>>4, c.regs.reg16(operand&0x0F))
		return 3
	})

	// Direct absolute addressing.
	setOp(0xD0, func(c *CPU) int { addr := uint32(c.fetch16()); c.regs.setReg8(1, c.bus.Read8(addr)); return 5 })
	setOp(0xD1, func(c *CPU) int { addr := uint32(c.fetch16()); c.bus.Write8(addr, c.regs.reg8(1)); return 5 })
	setOp(0xD2, func(c *CPU) int {
		addr := uint32(c.fetch16())
		c.regs.setReg16(3, c.bus.Read16(addr))
		return 6
	})
	setOp(0xD3, func(c *CPU) int {
		addr := uint32(c.fetch16())
		c.bus.Write16(addr, c.regs.reg16(3))
		return 6
	})
}
