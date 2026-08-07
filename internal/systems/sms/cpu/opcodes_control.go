package cpu

// decodeX3 handles the x=3 opcode block: everything involving the
// stack, control flow (JP/CALL/RET/RST), I/O, and the four prefix bytes.
func (c *CPU) decodeX3(y, z, p, q byte) int {
	switch z {
	case 0: // RET cc
		if c.condition(y) {
			c.regs.PC = c.pop()
			return 11
		}
		return 5
	case 1:
		return c.decodeX3Z1(p, q)
	case 2: // JP cc,nn
		target := c.fetch16()
		if c.condition(y) {
			c.regs.PC = target
		}
		return 10
	case 3:
		return c.decodeX3Z3(y)
	case 4: // CALL cc,nn
		target := c.fetch16()
		if c.condition(y) {
			c.push(c.regs.PC)
			c.regs.PC = target
			return 17
		}
		return 10
	case 5:
		return c.decodeX3Z5(p, q)
	case 6: // ALU A,n
		c.aluOp(y, c.fetchByte())
		return 7
	default: // RST y*8
		c.push(c.regs.PC)
		c.regs.PC = uint16(y) * 8
		return 11
	}
}

func (c *CPU) decodeX3Z1(p, q byte) int {
	if q == 0 {
		c.setRP2(p, c.pop())
		return 10
	}
	switch p {
	case 0: // RET
		c.regs.PC = c.pop()
		return 10
	case 1: // EXX
		c.regs.B, c.regs.B2 = c.regs.B2, c.regs.B
		c.regs.C, c.regs.C2 = c.regs.C2, c.regs.C
		c.regs.D, c.regs.D2 = c.regs.D2, c.regs.D
		c.regs.E, c.regs.E2 = c.regs.E2, c.regs.E
		c.regs.H, c.regs.H2 = c.regs.H2, c.regs.H
		c.regs.L, c.regs.L2 = c.regs.L2, c.regs.L
		return 4
	case 2: // JP (HL)
		c.regs.PC = c.regs.HL()
		return 4
	default: // LD SP,HL
		c.regs.SP = c.regs.HL()
		return 6
	}
}

func (c *CPU) decodeX3Z3(y byte) int {
	switch y {
	case 0: // JP nn
		c.regs.PC = c.fetch16()
		return 10
	case 1: // CB prefix
		return c.decodeCB()
	case 2: // OUT (n),A
		port := c.fetchByte()
		c.io.Out(port, c.regs.A)
		return 11
	case 3: // IN A,(n)
		port := c.fetchByte()
		c.regs.A = c.io.In(port)
		return 11
	case 4: // EX (SP),HL
		v := c.read16(c.regs.SP)
		c.write16(c.regs.SP, c.regs.HL())
		c.regs.SetHL(v)
		return 19
	case 5: // EX DE,HL
		de, hl := c.regs.DE(), c.regs.HL()
		c.regs.SetDE(hl)
		c.regs.SetHL(de)
		return 4
	case 6: // DI
		c.regs.IFF1, c.regs.IFF2 = false, false
		return 4
	default: // EI
		c.regs.IFF1, c.regs.IFF2 = true, true
		c.eiDelay = true
		return 4
	}
}

func (c *CPU) decodeX3Z5(p, q byte) int {
	if q == 0 {
		c.push(c.rp2(p))
		return 11
	}
	switch p {
	case 0: // CALL nn
		target := c.fetch16()
		c.push(c.regs.PC)
		c.regs.PC = target
		return 17
	case 1: // DD prefix
		return c.decodeIndexed(&c.regs.IX)
	case 2: // ED prefix
		return c.decodeED()
	default: // FD prefix
		return c.decodeIndexed(&c.regs.IY)
	}
}
