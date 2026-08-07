package cpu

// stepThumb decodes and executes one 16-bit Thumb instruction,
// dispatching by its top bits into the 19 documented Thumb instruction
// formats (see thumb_alu.go/thumb_branch.go/thumb_memory.go for each
// group's handlers).
func (c *CPU) stepThumb() int {
	op := c.fetch16()

	switch {
	case op&0xF800 == 0x1800: // format 2: add/subtract
		return c.thumbAddSub(op)
	case op&0xE000 == 0x0000: // format 1: move shifted register
		return c.thumbShift(op)
	case op&0xE000 == 0x2000: // format 3: move/compare/add/subtract immediate
		return c.thumbImmediate(op)
	case op&0xFC00 == 0x4000: // format 4: ALU operations
		return c.thumbALU(op)
	case op&0xFC00 == 0x4400: // format 5: hi register ops / BX
		return c.thumbHiReg(op)
	case op&0xF800 == 0x4800: // format 6: PC-relative load
		return c.thumbPCRelativeLoad(op)
	case op&0xF200 == 0x5000: // format 7: load/store with register offset
		return c.thumbLoadStoreReg(op)
	case op&0xF200 == 0x5200: // format 8: load/store sign-extended byte/halfword
		return c.thumbLoadStoreSignExt(op)
	case op&0xE000 == 0x6000: // format 9: load/store with immediate offset
		return c.thumbLoadStoreImm(op)
	case op&0xF000 == 0x8000: // format 10: load/store halfword
		return c.thumbLoadStoreHalf(op)
	case op&0xF000 == 0x9000: // format 11: SP-relative load/store
		return c.thumbSPRelative(op)
	case op&0xF000 == 0xA000: // format 12: load address
		return c.thumbLoadAddress(op)
	case op&0xFF00 == 0xB000: // format 13: add offset to SP
		return c.thumbAddSPOffset(op)
	case op&0xF600 == 0xB400: // format 14: push/pop registers
		return c.thumbPushPop(op)
	case op&0xF000 == 0xC000: // format 15: multiple load/store
		return c.thumbMultipleLoadStore(op)
	case op&0xFF00 == 0xDF00: // format 17: software interrupt
		return c.thumbSWI(op)
	case op&0xF000 == 0xD000: // format 16: conditional branch
		return c.thumbCondBranch(op)
	case op&0xF800 == 0xE000: // format 18: unconditional branch
		return c.thumbBranch(op)
	case op&0xF000 == 0xF000: // format 19: long branch with link
		return c.thumbBranchLink(op)
	default:
		return 1 // undefined encoding: treated as a 1-cycle NOP
	}
}

func (c *CPU) setNZ(v uint32) {
	c.regs.setFlag(FlagN, v&0x80000000 != 0)
	c.regs.setFlag(FlagZ, v == 0)
}
