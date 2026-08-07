package cpu

// armDataProcessing implements AND/EOR/SUB/RSB/ADD/ADC/SBC/RSC/TST/
// TEQ/CMP/CMN/ORR/MOV/BIC/MVN, all sharing the same Rd/Rn/operand2/S
// encoding.
func (c *CPU) armDataProcessing(op uint32) int {
	rn := (op >> 16) & 0x0F
	rd := (op >> 12) & 0x0F
	setFlags := op&0x00100000 != 0
	opcode := (op >> 21) & 0x0F

	a := c.regs.R[rn]
	b := c.operand2(op)

	var result uint32
	writesResult := true

	switch opcode {
	case 0x0: // AND
		result = a & b
	case 0x1: // EOR
		result = a ^ b
	case 0x2: // SUB
		result = c.doSub(a, b, setFlags)
	case 0x3: // RSB
		result = c.doSub(b, a, setFlags)
	case 0x4: // ADD
		result = c.doAdd(a, b, setFlags)
	case 0x5: // ADC
		carry := uint32(0)
		if c.regs.getFlag(FlagC) {
			carry = 1
		}
		result = c.doAdd(a, b+carry, setFlags)
	case 0x6: // SBC
		borrow := uint32(1)
		if c.regs.getFlag(FlagC) {
			borrow = 0
		}
		result = c.doSub(a, b+borrow, setFlags)
	case 0x7: // RSC
		borrow := uint32(1)
		if c.regs.getFlag(FlagC) {
			borrow = 0
		}
		result = c.doSub(b, a+borrow, setFlags)
	case 0x8: // TST
		result = a & b
		writesResult = false
	case 0x9: // TEQ
		result = a ^ b
		writesResult = false
	case 0xA: // CMP
		c.doSub(a, b, true)
		writesResult = false
	case 0xB: // CMN
		c.doAdd(a, b, true)
		writesResult = false
	case 0xC: // ORR
		result = a | b
	case 0xD: // MOV
		result = b
	case 0xE: // BIC
		result = a &^ b
	default: // MVN
		result = ^b
	}

	if !writesResult {
		if setFlags {
			c.setNZ(result)
		}
		return 1
	}

	arithmetic := opcode >= 0x2 && opcode <= 0x7
	if rd == 15 && setFlags {
		// The well-known exception-return idiom (e.g. "SUBS PC,LR,#4"):
		// writing R15 with the S bit set restores CPSR from SPSR instead
		// of updating flags normally.
		c.regs.R[15] = result &^ 3
		c.restoreFromSPSR()
		return 3
	}
	if setFlags && !arithmetic {
		c.setNZ(result)
	}
	c.regs.R[rd] = result
	if rd == 15 {
		c.regs.R[15] &^= 3
		return 3
	}
	return 1
}
