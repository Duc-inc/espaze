package spc700

func init() {
	setOp(0x00, func(c *CPU) int { return 2 }) // NOP
	setOp(0x01, func(c *CPU) int { c.stopped = true; return 2 })

	// LD reg,#imm8 / LD reg,(dp) / LD (dp),reg
	for r := byte(0); r < 3; r++ {
		reg := r
		setOp(0x10+reg, func(c *CPU) int {
			v := c.fetch8()
			*c.regPtr(reg) = v
			c.regs.setNZ(v)
			return 2
		})
		setOp(0x18+reg, func(c *CPU) int {
			v := c.read8(c.directPage())
			*c.regPtr(reg) = v
			c.regs.setNZ(v)
			return 3
		})
		setOp(0x1C+reg, func(c *CPU) int {
			c.write8(c.directPage(), *c.regPtr(reg))
			return 3
		})
	}

	setOp(0x20, func(c *CPU) int { v := c.read8(c.fetch16()); c.regs.A = v; c.regs.setNZ(v); return 4 })
	setOp(0x21, func(c *CPU) int { c.write8(c.fetch16(), c.regs.A); return 4 })

	setOp(0x28, func(c *CPU) int {
		operand := c.fetch8()
		v := *c.regPtr(operand & 0x03)
		*c.regPtr(operand >> 4 & 0x03) = v
		c.regs.setNZ(v)
		return 2
	})

	// ALU A,#imm8 and A,(dp): op = opcode&0x07 (ADD,ADC,SUB,SBC,CMP,AND,OR,XOR)
	for op := byte(0); op < 8; op++ {
		aluOp := op
		setOp(0x30+aluOp, func(c *CPU) int {
			imm := c.fetch8()
			c.regs.A = c.applyALU(aluOp, c.regs.A, imm)
			return 2
		})
		setOp(0x38+aluOp, func(c *CPU) int {
			v := c.read8(c.directPage())
			c.regs.A = c.applyALU(aluOp, c.regs.A, v)
			return 3
		})
	}

	for r := byte(0); r < 3; r++ {
		reg := r
		setOp(0x40+reg, func(c *CPU) int {
			p := c.regPtr(reg)
			*p++
			c.regs.setNZ(*p)
			return 2
		})
		setOp(0x44+reg, func(c *CPU) int {
			p := c.regPtr(reg)
			*p--
			c.regs.setNZ(*p)
			return 2
		})
	}
	setOp(0x48, func(c *CPU) int {
		addr := c.directPage()
		v := c.read8(addr) + 1
		c.write8(addr, v)
		c.regs.setNZ(v)
		return 4
	})
	setOp(0x49, func(c *CPU) int {
		addr := c.directPage()
		v := c.read8(addr) - 1
		c.write8(addr, v)
		c.regs.setNZ(v)
		return 4
	})

	for cc := byte(0); cc < 7; cc++ {
		cond := cc
		setOp(0x50+cond, func(c *CPU) int {
			offset := int8(c.fetch8())
			if c.checkCondition(cond) {
				c.regs.PC = uint16(int32(c.regs.PC) + int32(offset))
				return 4
			}
			return 2
		})
	}
	setOp(0x57, func(c *CPU) int {
		offset := int8(c.fetch8())
		c.regs.PC = uint16(int32(c.regs.PC) + int32(offset))
		return 4
	})

	setOp(0x60, func(c *CPU) int {
		target := c.fetch16()
		c.push16(c.regs.PC)
		c.regs.PC = target
		return 5
	})
	setOp(0x61, func(c *CPU) int { c.regs.PC = c.pop16(); return 4 })
	setOp(0x62, func(c *CPU) int {
		c.regs.PSW = c.pop8()
		c.regs.PC = c.pop16()
		return 5
	})

	for r := byte(0); r < 3; r++ {
		reg := r
		setOp(0x68+reg, func(c *CPU) int { c.push8(*c.regPtr(reg)); return 3 })
		setOp(0x6C+reg, func(c *CPU) int {
			v := c.pop8()
			*c.regPtr(reg) = v
			c.regs.setNZ(v)
			return 3
		})
	}
	setOp(0x6B, func(c *CPU) int { c.push8(c.regs.PSW); return 3 })
	setOp(0x6F, func(c *CPU) int { c.regs.PSW = c.pop8(); return 3 })

	setOp(0x70, func(c *CPU) int { c.regs.setFlag(FlagCarry, false); return 2 })
	setOp(0x71, func(c *CPU) int { c.regs.setFlag(FlagCarry, true); return 2 })
	setOp(0x74, func(c *CPU) int { c.regs.setFlag(FlagPage, false); return 2 })
	setOp(0x75, func(c *CPU) int { c.regs.setFlag(FlagPage, true); return 2 })
}
