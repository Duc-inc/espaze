package cpu

// The "A" 8-bit register is register code 1 (from XWA); "HL" the
// 16-bit register is code 3 (from XHL) - see registers.go's own
// numbering. Every 8-bit ALU op here targets A; every 16-bit ALU op
// targets HL, matching how the real chip's accumulator-style ALU
// instructions are conventionally used.

func init() {
	// ALU A,#imm8 (op = opcode&0x07: ADD,ADC,SUB,SBC,AND,XOR,OR,CP)
	for op := byte(0); op < 8; op++ {
		aluOp := op
		setOp(0x50+aluOp, func(c *CPU) int {
			imm := c.fetch8()
			result := c.applyALU8(aluOp, c.regs.reg8(1), imm)
			if aluOp != 7 {
				c.regs.setReg8(1, result)
			}
			return 3
		})
	}

	// ALU A,r8
	for op := byte(0); op < 8; op++ {
		aluOp := op
		setOp(0x58+aluOp, func(c *CPU) int {
			srcReg := c.fetch8() & 0x07
			result := c.applyALU8(aluOp, c.regs.reg8(1), c.regs.reg8(srcReg))
			if aluOp != 7 {
				c.regs.setReg8(1, result)
			}
			return 3
		})
	}

	// ALU HL,#imm16
	for op := byte(0); op < 8; op++ {
		aluOp := op
		setOp(0x60+aluOp, func(c *CPU) int {
			imm := c.fetch16()
			result := c.applyALU16(aluOp, c.regs.reg16(3), imm)
			if aluOp != 7 {
				c.regs.setReg16(3, result)
			}
			return 4
		})
	}

	// ALU HL,rp
	for op := byte(0); op < 8; op++ {
		aluOp := op
		setOp(0x68+aluOp, func(c *CPU) int {
			srcReg := c.fetch8() & 0x03
			result := c.applyALU16(aluOp, c.regs.reg16(3), c.regs.reg16(srcReg))
			if aluOp != 7 {
				c.regs.setReg16(3, result)
			}
			return 4
		})
	}

	// INC/DEC r8
	for r := byte(0); r < 8; r++ {
		reg := r
		setOp(0x70+reg, func(c *CPU) int {
			c.regs.setReg8(reg, c.add8(c.regs.reg8(reg), 1, false))
			return 2
		})
		setOp(0x78+reg, func(c *CPU) int {
			c.regs.setReg8(reg, c.sub8(c.regs.reg8(reg), 1, false))
			return 2
		})
	}

	// INC/DEC rp
	for p := byte(0); p < 4; p++ {
		pair := p
		setOp(0x80+pair, func(c *CPU) int {
			c.regs.setReg16(pair, c.regs.reg16(pair)+1)
			c.regs.setSZ16(c.regs.reg16(pair))
			return 3
		})
		setOp(0x84+pair, func(c *CPU) int {
			c.regs.setReg16(pair, c.regs.reg16(pair)-1)
			c.regs.setSZ16(c.regs.reg16(pair))
			return 3
		})
	}
}
