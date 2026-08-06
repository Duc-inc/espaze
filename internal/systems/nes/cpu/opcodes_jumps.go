package cpu

func opJMP(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.regs.PC = addr
	return 0
}

// opJSR pushes the address of the last byte of the JSR instruction
// itself (not the next instruction) - resolveOperand already advanced
// PC past the 2-byte target, so that's PC-1 at this point. RTS undoes
// it by popping and adding 1 back.
func opJSR(c *CPU, _ addrMode, addr uint16, _ bool) int {
	c.push16(c.regs.PC - 1)
	c.regs.PC = addr
	return 0
}

func opRTS(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.PC = c.pop16() + 1
	return 0
}

func opRTI(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.P = (c.pop() &^ FlagBreak) | FlagUnused
	c.regs.PC = c.pop16()
	return 0
}

func branch(c *CPU, addr uint16, pageCrossed, taken bool) int {
	if !taken {
		return 0
	}
	c.regs.PC = addr
	if pageCrossed {
		return 2
	}
	return 1
}

func opBCC(c *CPU, _ addrMode, addr uint16, pc bool) int {
	return branch(c, addr, pc, !c.regs.getFlag(FlagCarry))
}
func opBCS(c *CPU, _ addrMode, addr uint16, pc bool) int {
	return branch(c, addr, pc, c.regs.getFlag(FlagCarry))
}
func opBEQ(c *CPU, _ addrMode, addr uint16, pc bool) int {
	return branch(c, addr, pc, c.regs.getFlag(FlagZero))
}
func opBNE(c *CPU, _ addrMode, addr uint16, pc bool) int {
	return branch(c, addr, pc, !c.regs.getFlag(FlagZero))
}
func opBMI(c *CPU, _ addrMode, addr uint16, pc bool) int {
	return branch(c, addr, pc, c.regs.getFlag(FlagNegative))
}
func opBPL(c *CPU, _ addrMode, addr uint16, pc bool) int {
	return branch(c, addr, pc, !c.regs.getFlag(FlagNegative))
}
func opBVC(c *CPU, _ addrMode, addr uint16, pc bool) int {
	return branch(c, addr, pc, !c.regs.getFlag(FlagOverflow))
}
func opBVS(c *CPU, _ addrMode, addr uint16, pc bool) int {
	return branch(c, addr, pc, c.regs.getFlag(FlagOverflow))
}
