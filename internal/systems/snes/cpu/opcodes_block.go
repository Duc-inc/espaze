package cpu

// blockMove implements MVN/MVP: copies A+1 bytes from bank srcBank:X
// to bank destBank:Y, incrementing (MVN) or decrementing (MVP) both
// pointers each byte, leaving DBR set to the destination bank - all
// matching real hardware, except this project runs the whole transfer
// in one atomic Step call rather than the real chip's own
// interruptible-mid-transfer behavior (each byte re-executes the same
// instruction on real hardware, so an IRQ can land between bytes;
// this project's block transfers never do, the same simplification
// every other CPU core here makes for its own block-move instructions).
func (c *CPU) blockMove(forward bool) int {
	destBank := c.fetch8()
	srcBank := c.fetch8()

	count := uint32(c.regs.A) + 1
	for i := uint32(0); i < count; i++ {
		v := c.read8(uint32(srcBank)<<16 | uint32(c.regs.X))
		c.write8(uint32(destBank)<<16|uint32(c.regs.Y), v)
		if forward {
			c.regs.X++
			c.regs.Y++
		} else {
			c.regs.X--
			c.regs.Y--
		}
		c.regs.A--
	}
	c.regs.DBR = destBank
	return 7 + int(count)*7
}

func init() {
	setOp(0x54, func(c *CPU) int { return c.blockMove(true) })  // MVN
	setOp(0x44, func(c *CPU) int { return c.blockMove(false) }) // MVP
}
