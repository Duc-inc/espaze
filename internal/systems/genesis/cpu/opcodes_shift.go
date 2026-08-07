package cpu

// shiftOpcodes registers the register-destined forms of ASL/ASR, LSL/
// LSR and ROL/ROR (shifting a data register in place by an immediate
// 1-8 or a count from another Dn, mod 64). The memory-operand
// single-bit-shift forms and ROXL/ROXR (rotate through Extend) aren't
// implemented - real-world code overwhelmingly uses the register forms
// this covers.
func shiftOpcodes() []pattern {
	return []pattern{
		{mask: 0xF000, match: 0xE000, execute: shiftExecute},
	}
}

func shiftExecute(c *CPU, opcode uint16) int {
	sizeField := byte((opcode >> 6) & 0x03)
	if sizeField == 3 {
		return 4 // memory-operand form: not implemented, treated as a no-op
	}

	countField := byte((opcode >> 9) & 0x07)
	left := opcode&0x0100 != 0
	useRegister := opcode&0x0020 != 0
	kind := byte((opcode >> 3) & 0x03)
	reg := byte(opcode & 0x07)

	var count byte
	if useRegister {
		count = byte(c.regs.D[countField] % 64)
	} else {
		count = countField
		if count == 0 {
			count = 8
		}
	}

	value := maskSize(c.regs.D[reg], sizeField)
	bit := signBitFor(sizeField)

	var lastOut bool
	overflow := false
	carry := c.regs.getFlag(FlagC)

	for i := byte(0); i < count; i++ {
		switch kind {
		case 0: // arithmetic
			if left {
				signBefore := value&bit != 0
				lastOut = value&bit != 0
				value = maskSize(value<<1, sizeField)
				if value&bit != 0 != signBefore {
					overflow = true
				}
			} else {
				lastOut = value&1 != 0
				signBit := value & bit
				value = value>>1 | signBit
			}
		case 1: // logical
			if left {
				lastOut = value&bit != 0
				value = maskSize(value<<1, sizeField)
			} else {
				lastOut = value&1 != 0
				value >>= 1
			}
		default: // 3 = rotate (2, ROXx, isn't implemented - falls through as rotate too)
			if left {
				lastOut = value&bit != 0
				value = maskSize(value<<1, sizeField)
				if lastOut {
					value |= 1
				}
			} else {
				lastOut = value&1 != 0
				value >>= 1
				if lastOut {
					value |= bit
				}
			}
		}
		carry = lastOut
	}

	c.regs.D[reg] = mergeSize(c.regs.D[reg], value, sizeField)
	if count > 0 {
		c.regs.setFlag(FlagC, carry)
		if kind != 3 { // rotates (kind 3) don't touch X
			c.regs.setFlag(FlagX, carry)
		}
	}
	c.regs.setFlag(FlagN, isNegativeSized(value, sizeField))
	c.regs.setFlag(FlagZ, maskSize(value, sizeField) == 0)
	c.regs.setFlag(FlagV, kind == 0 && overflow)
	return 6 + int(count)*2
}
