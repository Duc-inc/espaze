package cpu

// opBRK triggers a software interrupt: it behaves like a 2-byte
// instruction (the byte after the opcode is skipped) even though it
// only occupies one byte in the program.
func opBRK(c *CPU, _ addrMode, _ uint16, _ bool) int {
	c.regs.PC++
	c.serviceInterrupt(irqVector, true)
	return 0
}

func opNOP(c *CPU, _ addrMode, _ uint16, _ bool) int { return 0 }
