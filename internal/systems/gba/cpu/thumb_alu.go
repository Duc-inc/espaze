package cpu

// thumbShift implements format 1: LSL/LSR/ASR Rd,Rs,#imm5.
func (c *CPU) thumbShift(op uint16) int {
	rd := op & 0x07
	rs := (op >> 3) & 0x07
	offset := uint32((op >> 6) & 0x1F)
	sub := (op >> 11) & 0x03

	v := c.regs.R[rs]
	var result uint32
	var carry bool
	switch sub {
	case 0: // LSL
		if offset == 0 {
			result, carry = v, c.regs.getFlag(FlagC)
		} else {
			result = v << offset
			carry = v&(1<<(32-offset)) != 0
		}
	case 1: // LSR
		if offset == 0 {
			result, carry = 0, v&0x80000000 != 0
		} else {
			result = v >> offset
			carry = v&(1<<(offset-1)) != 0
		}
	default: // ASR
		shift := offset
		if shift == 0 {
			shift = 31
		}
		result = uint32(int32(v) >> shift)
		carry = v&(1<<(shift-1)) != 0
	}

	c.regs.R[rd] = result
	c.setNZ(result)
	c.regs.setFlag(FlagC, carry)
	return 1
}

// thumbAddSub implements format 2: ADD/SUB Rd,Rs,Rn or Rd,Rs,#imm3.
func (c *CPU) thumbAddSub(op uint16) int {
	rd := op & 0x07
	rs := (op >> 3) & 0x07
	operand := uint32((op >> 6) & 0x07)
	if op&0x0400 == 0 { // register operand
		operand = c.regs.R[operand]
	}

	if op&0x0200 != 0 {
		c.regs.R[rd] = c.doSub(c.regs.R[rs], operand, true)
	} else {
		c.regs.R[rd] = c.doAdd(c.regs.R[rs], operand, true)
	}
	return 1
}

// thumbImmediate implements format 3: MOV/CMP/ADD/SUB Rd,#imm8.
func (c *CPU) thumbImmediate(op uint16) int {
	rd := (op >> 8) & 0x07
	imm := uint32(op & 0xFF)

	switch (op >> 11) & 0x03 {
	case 0: // MOV
		c.regs.R[rd] = imm
		c.setNZ(imm)
	case 1: // CMP
		c.doSub(c.regs.R[rd], imm, true)
	case 2: // ADD
		c.regs.R[rd] = c.doAdd(c.regs.R[rd], imm, true)
	default: // SUB
		c.regs.R[rd] = c.doSub(c.regs.R[rd], imm, true)
	}
	return 1
}

// thumbALU implements format 4: the 16 two-operand ALU operations.
func (c *CPU) thumbALU(op uint16) int {
	rd := op & 0x07
	rs := (op >> 3) & 0x07
	opcode := (op >> 6) & 0x0F
	a := c.regs.R[rd]
	b := c.regs.R[rs]

	switch opcode {
	case 0x0: // AND
		c.regs.R[rd] = a & b
		c.setNZ(c.regs.R[rd])
	case 0x1: // EOR
		c.regs.R[rd] = a ^ b
		c.setNZ(c.regs.R[rd])
	case 0x2: // LSL (by register)
		c.regs.R[rd] = shiftLeftReg(a, b, &c.regs.CPSR)
		c.setNZ(c.regs.R[rd])
	case 0x3: // LSR
		c.regs.R[rd] = shiftRightReg(a, b, &c.regs.CPSR)
		c.setNZ(c.regs.R[rd])
	case 0x4: // ASR
		c.regs.R[rd] = arithShiftRightReg(a, b, &c.regs.CPSR)
		c.setNZ(c.regs.R[rd])
	case 0x5: // ADC
		carry := uint32(0)
		if c.regs.getFlag(FlagC) {
			carry = 1
		}
		c.regs.R[rd] = c.doAdd(a, b+carry, true)
	case 0x6: // SBC
		borrow := uint32(1)
		if c.regs.getFlag(FlagC) {
			borrow = 0
		}
		c.regs.R[rd] = c.doSub(a, b+borrow, true)
	case 0x7: // ROR (by register)
		c.regs.R[rd] = rotateRightReg(a, b, &c.regs.CPSR)
		c.setNZ(c.regs.R[rd])
	case 0x8: // TST
		c.setNZ(a & b)
	case 0x9: // NEG
		c.regs.R[rd] = c.doSub(0, b, true)
	case 0xA: // CMP
		c.doSub(a, b, true)
	case 0xB: // CMN
		c.doAdd(a, b, true)
	case 0xC: // ORR
		c.regs.R[rd] = a | b
		c.setNZ(c.regs.R[rd])
	case 0xD: // MUL
		c.regs.R[rd] = a * b
		c.setNZ(c.regs.R[rd])
	case 0xE: // BIC
		c.regs.R[rd] = a &^ b
		c.setNZ(c.regs.R[rd])
	default: // MVN
		c.regs.R[rd] = ^b
		c.setNZ(c.regs.R[rd])
	}
	return 1
}
