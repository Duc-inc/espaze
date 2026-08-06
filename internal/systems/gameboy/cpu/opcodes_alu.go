package cpu

// aluOps lists the 8 accumulator operations in the order the opcode
// encoding uses them, both for 0x80-0xBF (A,r8) and 0xC6-0xFE (A,d8).
var aluOps = [8]func(c *CPU, v byte){
	func(c *CPU, v byte) { c.regs.A = c.add8(c.regs.A, v) },
	func(c *CPU, v byte) { c.regs.A = c.adc8(c.regs.A, v) },
	func(c *CPU, v byte) { c.regs.A = c.sub8(c.regs.A, v) },
	func(c *CPU, v byte) { c.regs.A = c.sbc8(c.regs.A, v) },
	func(c *CPU, v byte) { c.regs.A = c.and8(c.regs.A, v) },
	func(c *CPU, v byte) { c.regs.A = c.xor8(c.regs.A, v) },
	func(c *CPU, v byte) { c.regs.A = c.or8(c.regs.A, v) },
	func(c *CPU, v byte) { c.cp8(c.regs.A, v) },
}

func init() {
	// 0x80-0xBF: <op> A,r8
	for opIdx := byte(0); opIdx < 8; opIdx++ {
		for r := byte(0); r < 8; r++ {
			opcode := 0x80 + opIdx*8 + r
			op, reg := aluOps[opIdx], r
			cycles := 4
			if reg == 6 {
				cycles = 8
			}
			mainTable[opcode] = func(c *CPU) int {
				op(c, c.readR8(reg))
				return cycles
			}
		}
	}

	// 0xC6,0xCE,...,0xFE: <op> A,d8
	for opIdx := byte(0); opIdx < 8; opIdx++ {
		op := aluOps[opIdx]
		mainTable[0xC6+opIdx*8] = func(c *CPU) int {
			op(c, c.fetch8())
			return 8
		}
	}

	// INC/DEC r8
	for idx := byte(0); idx < 8; idx++ {
		i := idx
		cycles := 4
		if i == 6 {
			cycles = 12
		}
		mainTable[0x04+i*8] = func(c *CPU) int { c.writeR8(i, c.inc8(c.readR8(i))); return cycles }
		mainTable[0x05+i*8] = func(c *CPU) int { c.writeR8(i, c.dec8(c.readR8(i))); return cycles }
	}

	// INC/DEC rr, ADD HL,rr
	for idx := byte(0); idx < 4; idx++ {
		i := idx
		mainTable[0x03+i*0x10] = func(c *CPU) int { c.writeR16(i, c.readR16(i)+1); return 8 }
		mainTable[0x0B+i*0x10] = func(c *CPU) int { c.writeR16(i, c.readR16(i)-1); return 8 }
		mainTable[0x09+i*0x10] = func(c *CPU) int { c.addHL(c.readR16(i)); return 8 }
	}

	mainTable[0xE8] = func(c *CPU) int {
		c.regs.SP = c.addSPSigned(int8(c.fetch8()))
		return 16
	}
}

func init() {
	mainTable[0x07] = func(c *CPU) int { c.regs.A = c.rlc(c.regs.A); c.regs.SetFlag(FlagZ, false); return 4 }
	mainTable[0x0F] = func(c *CPU) int { c.regs.A = c.rrc(c.regs.A); c.regs.SetFlag(FlagZ, false); return 4 }
	mainTable[0x17] = func(c *CPU) int { c.regs.A = c.rl(c.regs.A); c.regs.SetFlag(FlagZ, false); return 4 }
	mainTable[0x1F] = func(c *CPU) int { c.regs.A = c.rr(c.regs.A); c.regs.SetFlag(FlagZ, false); return 4 }

	mainTable[0x27] = func(c *CPU) int { c.daa(); return 4 }
	mainTable[0x2F] = func(c *CPU) int {
		c.regs.A = ^c.regs.A
		c.regs.SetFlag(FlagN, true)
		c.regs.SetFlag(FlagH, true)
		return 4
	}
	mainTable[0x37] = func(c *CPU) int {
		c.regs.SetFlag(FlagN, false)
		c.regs.SetFlag(FlagH, false)
		c.regs.SetFlag(FlagC, true)
		return 4
	}
	mainTable[0x3F] = func(c *CPU) int {
		c.regs.SetFlag(FlagN, false)
		c.regs.SetFlag(FlagH, false)
		c.regs.SetFlag(FlagC, !c.regs.HasFlag(FlagC))
		return 4
	}
}

// daa implements DAA: adjusts A after a BCD add/sub so it holds two valid
// packed decimal digits again, per the standard LR35902 algorithm.
func (c *CPU) daa() {
	a := c.regs.A
	adjust := byte(0)
	carry := false

	if c.regs.HasFlag(FlagN) {
		if c.regs.HasFlag(FlagH) {
			adjust += 0x06
		}
		if c.regs.HasFlag(FlagC) {
			adjust += 0x60
			carry = true
		}
		a -= adjust
	} else {
		if c.regs.HasFlag(FlagH) || a&0x0F > 0x09 {
			adjust += 0x06
		}
		if c.regs.HasFlag(FlagC) || a > 0x99 {
			adjust += 0x60
			carry = true
		}
		a += adjust
	}

	c.regs.A = a
	c.regs.SetFlag(FlagZ, a == 0)
	c.regs.SetFlag(FlagH, false)
	c.regs.SetFlag(FlagC, carry)
}
