package cpu

// conditions lists the four branch conditions in the order the opcode
// encoding uses them for JR/JP/CALL/RET: NZ, Z, NC, C.
var conditions = [4]func(c *CPU) bool{
	func(c *CPU) bool { return !c.regs.HasFlag(FlagZ) },
	func(c *CPU) bool { return c.regs.HasFlag(FlagZ) },
	func(c *CPU) bool { return !c.regs.HasFlag(FlagC) },
	func(c *CPU) bool { return c.regs.HasFlag(FlagC) },
}

func init() {
	mainTable[0x00] = func(c *CPU) int { return 4 }
	mainTable[0x10] = func(c *CPU) int { c.fetch8(); c.stopped = true; return 4 }
	mainTable[0x76] = func(c *CPU) int { c.halted = true; return 4 }
	mainTable[0xF3] = func(c *CPU) int { c.ime, c.eiDelay = false, 0; return 4 }
	mainTable[0xFB] = func(c *CPU) int { c.eiDelay = 2; return 4 }

	// JR r8 / JR cc,r8
	mainTable[0x18] = func(c *CPU) int { c.jumpRelative(); return 12 }
	for idx := byte(0); idx < 4; idx++ {
		cond := conditions[idx]
		mainTable[0x20+idx*8] = func(c *CPU) int {
			if cond(c) {
				c.jumpRelative()
				return 12
			}
			c.fetch8()
			return 8
		}
	}

	// JP a16 / JP cc,a16 / JP (HL)
	mainTable[0xC3] = func(c *CPU) int { c.regs.PC = c.fetch16(); return 16 }
	mainTable[0xE9] = func(c *CPU) int { c.regs.PC = c.regs.HL(); return 4 }
	for idx := byte(0); idx < 4; idx++ {
		cond := conditions[idx]
		mainTable[0xC2+idx*8] = func(c *CPU) int {
			addr := c.fetch16()
			if cond(c) {
				c.regs.PC = addr
				return 16
			}
			return 12
		}
	}

	// CALL a16 / CALL cc,a16
	mainTable[0xCD] = func(c *CPU) int {
		addr := c.fetch16()
		c.push16(c.regs.PC)
		c.regs.PC = addr
		return 24
	}
	for idx := byte(0); idx < 4; idx++ {
		cond := conditions[idx]
		mainTable[0xC4+idx*8] = func(c *CPU) int {
			addr := c.fetch16()
			if cond(c) {
				c.push16(c.regs.PC)
				c.regs.PC = addr
				return 24
			}
			return 12
		}
	}

	// RET / RET cc / RETI
	mainTable[0xC9] = func(c *CPU) int { c.regs.PC = c.pop16(); return 16 }
	mainTable[0xD9] = func(c *CPU) int { c.regs.PC = c.pop16(); c.ime = true; return 16 }
	for idx := byte(0); idx < 4; idx++ {
		cond := conditions[idx]
		mainTable[0xC0+idx*8] = func(c *CPU) int {
			if cond(c) {
				c.regs.PC = c.pop16()
				return 20
			}
			return 8
		}
	}

	// RST n
	for idx := byte(0); idx < 8; idx++ {
		target := uint16(idx) * 8
		mainTable[0xC7+idx*8] = func(c *CPU) int {
			c.push16(c.regs.PC)
			c.regs.PC = target
			return 16
		}
	}

	// Illegal/undefined opcodes: real hardware locks up, we just no-op
	// so a game that (incorrectly) hits one doesn't take the whole app down.
	for _, illegal := range []byte{0xD3, 0xDB, 0xDD, 0xE3, 0xE4, 0xEB, 0xEC, 0xED, 0xF4, 0xFC, 0xFD} {
		mainTable[illegal] = func(c *CPU) int { return 4 }
	}
}

func (c *CPU) jumpRelative() {
	offset := int8(c.fetch8())
	c.regs.PC = uint16(int32(c.regs.PC) + int32(offset))
}
