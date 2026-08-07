package cpu

// singleTransferOffset decodes a single-data-transfer instruction's
// 12-bit offset field: an immediate, or a register optionally shifted
// by an immediate amount (the register-specified-shift-amount form
// isn't valid in this instruction class on real hardware either).
func (c *CPU) singleTransferOffset(op uint32) uint32 {
	if op&0x02000000 == 0 {
		return op & 0xFFF
	}
	rm := c.regs.R[op&0x0F]
	amount := (op >> 7) & 0x1F
	switch (op >> 5) & 0x03 {
	case 0:
		return shiftLeftReg(rm, amount, &c.regs.CPSR)
	case 1:
		if amount == 0 {
			amount = 32
		}
		return shiftRightReg(rm, amount, &c.regs.CPSR)
	case 2:
		if amount == 0 {
			amount = 32
		}
		return arithShiftRightReg(rm, amount, &c.regs.CPSR)
	default:
		return rotateRightReg(rm, amount, &c.regs.CPSR)
	}
}

// armSingleTransfer implements LDR/STR{B}.
func (c *CPU) armSingleTransfer(op uint32) int {
	rn := (op >> 16) & 0x0F
	rd := (op >> 12) & 0x0F
	pre := op&0x01000000 != 0
	up := op&0x00800000 != 0
	byteOp := op&0x00400000 != 0
	writeback := op&0x00200000 != 0
	load := op&0x00100000 != 0

	offset := c.singleTransferOffset(op)
	base := c.regs.R[rn]
	addr := base
	if pre {
		if up {
			addr = base + offset
		} else {
			addr = base - offset
		}
	}

	if load {
		if byteOp {
			c.regs.R[rd] = uint32(c.read8(addr))
		} else {
			c.regs.R[rd] = c.read32(addr)
		}
	} else {
		if byteOp {
			c.write8(addr, byte(c.regs.R[rd]))
		} else {
			c.write32(addr, c.regs.R[rd])
		}
	}

	final := addr
	if !pre {
		if up {
			final = base + offset
		} else {
			final = base - offset
		}
	}
	if writeback || !pre {
		c.regs.R[rn] = final
	}
	return 3
}

// armHalfwordTransfer implements LDRH/STRH/LDRSB/LDRSH.
func (c *CPU) armHalfwordTransfer(op uint32) int {
	rn := (op >> 16) & 0x0F
	rd := (op >> 12) & 0x0F
	pre := op&0x01000000 != 0
	up := op&0x00800000 != 0
	immOffset := op&0x00400000 != 0
	writeback := op&0x00200000 != 0
	load := op&0x00100000 != 0

	var offset uint32
	if immOffset {
		offset = (op>>8)&0x0F<<4 | op&0x0F
	} else {
		offset = c.regs.R[op&0x0F]
	}

	base := c.regs.R[rn]
	addr := base
	if pre {
		if up {
			addr = base + offset
		} else {
			addr = base - offset
		}
	}

	sh := (op >> 5) & 0x03
	if load {
		switch sh {
		case 1:
			c.regs.R[rd] = uint32(c.read16(addr))
		case 2:
			c.regs.R[rd] = uint32(int32(int8(c.read8(addr))))
		default:
			c.regs.R[rd] = uint32(int32(int16(c.read16(addr))))
		}
	} else {
		c.write16(addr, uint16(c.regs.R[rd]))
	}

	if !pre {
		if up {
			c.regs.R[rn] = base + offset
		} else {
			c.regs.R[rn] = base - offset
		}
	} else if writeback {
		c.regs.R[rn] = addr
	}
	return 3
}

// armBlockTransfer implements LDM/STM. The "S bit" (user-bank
// register access / force-user-mode) and the edge case of Rn itself
// appearing in the register list aren't specially handled - a minor
// simplification consistent with the rest of this decoder's scope.
func (c *CPU) armBlockTransfer(op uint32) int {
	rn := (op >> 16) & 0x0F
	pre := op&0x01000000 != 0
	up := op&0x00800000 != 0
	writeback := op&0x00200000 != 0
	load := op&0x00100000 != 0
	list := uint16(op)

	step := int32(4)
	if !up {
		step = -4
	}
	addr := int32(c.regs.R[rn])

	for r := 0; r < 16; r++ {
		if list&(1<<uint(r)) == 0 {
			continue
		}
		if pre {
			addr += step
		}
		if load {
			c.regs.R[r] = c.read32(uint32(addr))
		} else {
			c.write32(uint32(addr), c.regs.R[r])
		}
		if !pre {
			addr += step
		}
	}
	if writeback {
		c.regs.R[rn] = uint32(addr)
	}
	return 3
}
