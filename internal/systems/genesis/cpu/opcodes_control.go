package cpu

// controlOpcodes registers branches, jumps/subroutine calls/returns,
// DBcc/Scc, MOVEM, and TRAP. DBcc is registered before Scc: DBcc's
// exact bit pattern lands inside a slot Scc's own destination-mode
// field would otherwise treat as "target an address register", which
// Scc can never actually do - so the more specific DBcc pattern has to
// win first.
func controlOpcodes() []pattern {
	return []pattern{
		{mask: 0xF000, match: 0x6000, execute: bccExecute},
		{mask: 0xF0F8, match: 0x50C8, execute: dbccExecute},
		{mask: 0xF0C0, match: 0x50C0, execute: sccExecute},
		{mask: 0xFFC0, match: 0x4EC0, execute: jmpExecute},
		{mask: 0xFFC0, match: 0x4E80, execute: jsrExecute},
		{mask: 0xFFFF, match: 0x4E75, execute: rtsExecute},
		{mask: 0xFFFF, match: 0x4E73, execute: rteExecute},
		{mask: 0xFFFF, match: 0x4E77, execute: rtrExecute},
		{mask: 0xFFF0, match: 0x4E40, execute: trapExecute},
		{mask: 0xFF80, match: 0x4880, execute: movemExecute}, // store (dr=0)
		{mask: 0xFF80, match: 0x4C80, execute: movemExecute}, // load (dr=1)
	}
}

func bccExecute(c *CPU, opcode uint16) int {
	cc := byte((opcode >> 8) & 0x0F)
	disp8 := int8(opcode)
	base := c.regs.PC

	var disp int32
	if disp8 == 0 {
		disp = int32(int16(c.fetchWord()))
	} else {
		disp = int32(disp8)
	}
	target := uint32(int32(base) + disp)

	switch cc {
	case 0x0: // BRA
		c.regs.PC = target
		return 10
	case 0x1: // BSR
		c.push32(c.regs.PC)
		c.regs.PC = target
		return 18
	default:
		if c.evalCondition(cc) {
			c.regs.PC = target
			return 10
		}
		return 8
	}
}

func dbccExecute(c *CPU, opcode uint16) int {
	cc := byte((opcode >> 8) & 0x0F)
	reg := byte(opcode & 0x07)
	base := c.regs.PC
	disp := int32(int16(c.fetchWord()))

	if c.evalCondition(cc) {
		return 12
	}
	count := int16(c.regs.D[reg]) - 1
	c.regs.D[reg] = mergeSize(c.regs.D[reg], uint32(uint16(count)), sizeWord)
	if count != -1 {
		c.regs.PC = uint32(int32(base) + disp)
		return 10
	}
	return 14
}

func sccExecute(c *CPU, opcode uint16) int {
	cc := byte((opcode >> 8) & 0x0F)
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)
	loc := c.resolveEA(mode, eaReg, sizeByte)

	if c.evalCondition(cc) {
		c.writeLocation(loc, sizeByte, 0xFF)
		return 6
	}
	c.writeLocation(loc, sizeByte, 0x00)
	return 4
}

func jmpExecute(c *CPU, opcode uint16) int {
	mode := byte((opcode >> 3) & 0x07)
	reg := byte(opcode & 0x07)
	c.regs.PC = c.resolveEA(mode, reg, sizeLong).addr
	return 8
}

func jsrExecute(c *CPU, opcode uint16) int {
	mode := byte((opcode >> 3) & 0x07)
	reg := byte(opcode & 0x07)
	target := c.resolveEA(mode, reg, sizeLong).addr
	c.push32(c.regs.PC)
	c.regs.PC = target
	return 16
}

func rtsExecute(c *CPU, opcode uint16) int {
	c.regs.PC = c.pop32()
	return 16
}

// rteExecute implements RTE: restores SR then PC. Real hardware
// privilege-checks this (supervisor-only); this project trusts the ROM
// code it runs, the same simplification every other core here makes
// for its own privileged/system instructions.
func rteExecute(c *CPU, opcode uint16) int {
	sr := c.bus.Read16(c.regs.A[7])
	c.regs.A[7] += 2
	c.regs.SR = sr
	c.regs.PC = c.pop32()
	return 20
}

// rtrExecute implements RTR: like RTS, but also restores the CCR bits
// (not the privileged system byte) from the stack.
func rtrExecute(c *CPU, opcode uint16) int {
	ccr := c.bus.Read16(c.regs.A[7])
	c.regs.A[7] += 2
	c.regs.SR = c.regs.SR&0xFF00 | ccr&0x00FF
	c.regs.PC = c.pop32()
	return 20
}

func trapExecute(c *CPU, opcode uint16) int {
	vector := uint32(vectorTrap0) + uint32(opcode&0x0F)
	return c.raiseException(vector)
}

// movemExecute implements MOVEM: transfers any subset of D0-D7/A0-A7
// to or from memory. The predecrement addressing mode uses a *reversed*
// register-mask bit order from every other mode - a real hardware quirk
// (see movemOrder) that exists so the registers land in memory in the
// same order regardless of which direction the stack grows.
func movemExecute(c *CPU, opcode uint16) int {
	toRegs := opcode&0x0400 != 0
	long := opcode&0x0040 != 0
	mode := byte((opcode >> 3) & 0x07)
	eaReg := byte(opcode & 0x07)
	size := sizeWord
	if long {
		size = sizeLong
	}
	regMask := c.fetchWord()
	step := uint32(2)
	if long {
		step = 4
	}

	order := movemOrder(mode == 4)
	count := 0

	if mode == 4 { // -(An): store only, reversed order, address decrements first each time
		addr := c.regs.A[eaReg]
		for i := 0; i < 16; i++ {
			if regMask&(1<<uint(i)) == 0 {
				continue
			}
			addr -= step
			writeMovemReg(c, addr, size, order[i])
			count++
		}
		c.regs.A[eaReg] = addr
		return 8 + count*4
	}

	loc := c.resolveEA(mode, eaReg, size)
	addr := loc.addr
	for i := 0; i < 16; i++ {
		if regMask&(1<<uint(i)) == 0 {
			continue
		}
		if toRegs {
			readMovemReg(c, addr, size, order[i])
		} else {
			writeMovemReg(c, addr, size, order[i])
		}
		addr += step
		count++
	}
	if mode == 3 { // (An)+ also writes the final address back
		c.regs.A[eaReg] = addr
	}
	return 8 + count*4
}

type movemReg struct {
	isAddr bool
	num    byte
}

func movemOrder(predecrement bool) [16]movemReg {
	var order [16]movemReg
	if predecrement {
		for i := 0; i < 8; i++ {
			order[i] = movemReg{isAddr: true, num: byte(7 - i)}
			order[8+i] = movemReg{isAddr: false, num: byte(7 - i)}
		}
	} else {
		for i := 0; i < 8; i++ {
			order[i] = movemReg{isAddr: false, num: byte(i)}
			order[8+i] = movemReg{isAddr: true, num: byte(i)}
		}
	}
	return order
}

func writeMovemReg(c *CPU, addr uint32, size byte, r movemReg) {
	var v uint32
	if r.isAddr {
		v = c.regs.A[r.num]
	} else {
		v = c.regs.D[r.num]
	}
	if size == sizeLong {
		c.bus.Write32(addr, v)
	} else {
		c.bus.Write16(addr, uint16(v))
	}
}

func readMovemReg(c *CPU, addr uint32, size byte, r movemReg) {
	var v uint32
	if size == sizeLong {
		v = c.bus.Read32(addr)
	} else {
		v = uint32(int32(int16(c.bus.Read16(addr))))
	}
	if r.isAddr {
		c.regs.A[r.num] = v
	} else {
		c.regs.D[r.num] = v
	}
}
