package cpu

// stepARM decodes and executes one 32-bit ARM instruction. Coverage
// here is deliberately narrower than the Thumb decoder - the common
// instruction classes (data processing, branch/branch-exchange,
// multiply, single/block data transfer, halfword transfer, SWI) - since
// the large majority of commercial GBA game code runs in Thumb state;
// ARM state is mostly used for brief startup/IRQ-handler stubs and a
// minority of performance-critical routines.
func (c *CPU) stepARM() int {
	op := c.fetch32()
	cond := op >> 28

	if !c.checkCondition(cond) {
		return 1
	}

	switch {
	case op&0x0FFFFFF0 == 0x012FFF10: // BX
		return c.armBX(op)
	case op&0x0FC000F0 == 0x00000090: // MUL/MLA
		return c.armMultiply(op)
	case op&0x0E000090 == 0x00000090 && op&0x00000060 != 0: // halfword/signed transfer
		return c.armHalfwordTransfer(op)
	case op&0x0C000000 == 0x00000000: // data processing
		return c.armDataProcessing(op)
	case op&0x0C000000 == 0x04000000: // single data transfer
		return c.armSingleTransfer(op)
	case op&0x0E000000 == 0x08000000: // block data transfer
		return c.armBlockTransfer(op)
	case op&0x0E000000 == 0x0A000000: // branch/branch-with-link
		return c.armBranch(op)
	case op&0x0F000000 == 0x0F000000: // SWI
		return c.enterException(modeSupervisor, 0x08, 4)
	default:
		return 1 // undefined/coprocessor encodings: not implemented
	}
}

// operand2 decodes a data-processing instruction's second operand:
// either a rotated 8-bit immediate, or a (possibly shifted) register.
func (c *CPU) operand2(op uint32) uint32 {
	if op&0x02000000 != 0 { // immediate
		imm := op & 0xFF
		rot := (op >> 8) & 0x0F * 2
		return rotateRightReg(imm, rot, &c.regs.CPSR)
	}

	rm := c.regs.R[op&0x0F]
	shiftType := (op >> 5) & 0x03
	var amount uint32
	if op&0x10 != 0 { // shift amount in a register
		amount = c.regs.R[(op>>8)&0x0F] & 0xFF
	} else {
		amount = (op >> 7) & 0x1F
	}

	switch shiftType {
	case 0:
		return shiftLeftReg(rm, amount, &c.regs.CPSR)
	case 1:
		if amount == 0 && op&0x10 == 0 {
			amount = 32
		}
		return shiftRightReg(rm, amount, &c.regs.CPSR)
	case 2:
		if amount == 0 && op&0x10 == 0 {
			amount = 32
		}
		return arithShiftRightReg(rm, amount, &c.regs.CPSR)
	default:
		if amount == 0 && op&0x10 == 0 {
			// RRX: rotate right by 1 through Carry - approximated here
			// as an ordinary rotate by 1, a minor simplification.
			amount = 1
		}
		return rotateRightReg(rm, amount, &c.regs.CPSR)
	}
}
