package cpu

// thumbLoadStoreReg implements format 7: LDR/STR{B} Rd,[Rb,Ro].
func (c *CPU) thumbLoadStoreReg(op uint16) int {
	rd := op & 0x07
	rb := (op >> 3) & 0x07
	ro := (op >> 6) & 0x07
	addr := c.regs.R[rb] + c.regs.R[ro]

	load := op&0x0800 != 0
	byteOp := op&0x0400 != 0
	switch {
	case load && byteOp:
		c.regs.R[rd] = uint32(c.read8(addr))
	case load:
		c.regs.R[rd] = c.read32(addr)
	case byteOp:
		c.write8(addr, byte(c.regs.R[rd]))
	default:
		c.write32(addr, c.regs.R[rd])
	}
	return 2
}

// thumbLoadStoreSignExt implements format 8: STRH/LDRH/LDSB/LDSH.
func (c *CPU) thumbLoadStoreSignExt(op uint16) int {
	rd := op & 0x07
	rb := (op >> 3) & 0x07
	ro := (op >> 6) & 0x07
	addr := c.regs.R[rb] + c.regs.R[ro]

	h := op&0x0800 != 0
	s := op&0x0400 != 0
	switch {
	case !s && !h: // STRH
		c.write16(addr, uint16(c.regs.R[rd]))
	case !s && h: // LDRH
		c.regs.R[rd] = uint32(c.read16(addr))
	case s && !h: // LDSB
		c.regs.R[rd] = uint32(int32(int8(c.read8(addr))))
	default: // LDSH
		c.regs.R[rd] = uint32(int32(int16(c.read16(addr))))
	}
	return 2
}

// thumbLoadStoreImm implements format 9.
func (c *CPU) thumbLoadStoreImm(op uint16) int {
	rd := op & 0x07
	rb := (op >> 3) & 0x07
	offset := uint32((op >> 6) & 0x1F)
	byteOp := op&0x1000 != 0
	if !byteOp {
		offset *= 4
	}
	addr := c.regs.R[rb] + offset

	load := op&0x0800 != 0
	switch {
	case load && byteOp:
		c.regs.R[rd] = uint32(c.read8(addr))
	case load:
		c.regs.R[rd] = c.read32(addr)
	case byteOp:
		c.write8(addr, byte(c.regs.R[rd]))
	default:
		c.write32(addr, c.regs.R[rd])
	}
	return 2
}

// thumbLoadStoreHalf implements format 10.
func (c *CPU) thumbLoadStoreHalf(op uint16) int {
	rd := op & 0x07
	rb := (op >> 3) & 0x07
	offset := uint32((op>>6)&0x1F) * 2
	addr := c.regs.R[rb] + offset

	if op&0x0800 != 0 {
		c.regs.R[rd] = uint32(c.read16(addr))
	} else {
		c.write16(addr, uint16(c.regs.R[rd]))
	}
	return 2
}

// thumbSPRelative implements format 11.
func (c *CPU) thumbSPRelative(op uint16) int {
	rd := (op >> 8) & 0x07
	offset := uint32(op&0xFF) * 4
	addr := c.regs.R[13] + offset

	if op&0x0800 != 0 {
		c.regs.R[rd] = c.read32(addr)
	} else {
		c.write32(addr, c.regs.R[rd])
	}
	return 2
}

// thumbLoadAddress implements format 12: ADD Rd,PC/SP,#imm8*4.
func (c *CPU) thumbLoadAddress(op uint16) int {
	rd := (op >> 8) & 0x07
	offset := uint32(op&0xFF) * 4
	if op&0x0800 != 0 {
		c.regs.R[rd] = c.regs.R[13] + offset
	} else {
		c.regs.R[rd] = (c.regs.R[15]+2)&^3 + offset
	}
	return 1
}

// thumbAddSPOffset implements format 13.
func (c *CPU) thumbAddSPOffset(op uint16) int {
	offset := uint32(op&0x7F) * 4
	if op&0x80 != 0 {
		c.regs.R[13] -= offset
	} else {
		c.regs.R[13] += offset
	}
	return 1
}

// thumbPushPop implements format 14.
func (c *CPU) thumbPushPop(op uint16) int {
	list := byte(op)
	pop := op&0x0800 != 0
	includesExtra := op&0x0100 != 0

	if pop {
		for r := 0; r < 8; r++ {
			if list&(1<<uint(r)) != 0 {
				c.regs.R[r] = c.pop()
			}
		}
		if includesExtra {
			c.regs.R[15] = c.pop() &^ 1
		}
	} else {
		if includesExtra {
			c.push(c.regs.R[14])
		}
		for r := 7; r >= 0; r-- {
			if list&(1<<uint(r)) != 0 {
				c.push(c.regs.R[r])
			}
		}
	}
	return 3
}

// thumbMultipleLoadStore implements format 15: LDMIA/STMIA Rb!,{list}.
func (c *CPU) thumbMultipleLoadStore(op uint16) int {
	rb := (op >> 8) & 0x07
	list := byte(op)
	load := op&0x0800 != 0
	addr := c.regs.R[rb]

	for r := 0; r < 8; r++ {
		if list&(1<<uint(r)) == 0 {
			continue
		}
		if load {
			c.regs.R[r] = c.read32(addr)
		} else {
			c.write32(addr, c.regs.R[r])
		}
		addr += 4
	}
	c.regs.R[rb] = addr
	return 3
}
