package cpu

// Each shift/rotate works on either the accumulator or a memory operand
// depending on addressing mode, so they share this read/modify/write
// helper instead of duplicating the mode check four times over.
func shiftRMW(c *CPU, mode addrMode, addr uint16, f func(byte) byte) {
	var value byte
	if mode == modeAccumulator {
		value = c.regs.A
	} else {
		value = c.bus.Read(addr)
	}

	result := f(value)

	if mode == modeAccumulator {
		c.regs.A = result
	} else {
		c.bus.Write(addr, result)
	}
	c.regs.setZN(result)
}

func opASL(c *CPU, mode addrMode, addr uint16, _ bool) int {
	shiftRMW(c, mode, addr, func(v byte) byte {
		c.regs.setFlag(FlagCarry, v&0x80 != 0)
		return v << 1
	})
	return 0
}

func opLSR(c *CPU, mode addrMode, addr uint16, _ bool) int {
	shiftRMW(c, mode, addr, func(v byte) byte {
		c.regs.setFlag(FlagCarry, v&0x01 != 0)
		return v >> 1
	})
	return 0
}

func opROL(c *CPU, mode addrMode, addr uint16, _ bool) int {
	shiftRMW(c, mode, addr, func(v byte) byte {
		carryIn := byte(0)
		if c.regs.getFlag(FlagCarry) {
			carryIn = 1
		}
		c.regs.setFlag(FlagCarry, v&0x80 != 0)
		return v<<1 | carryIn
	})
	return 0
}

func opROR(c *CPU, mode addrMode, addr uint16, _ bool) int {
	shiftRMW(c, mode, addr, func(v byte) byte {
		carryIn := byte(0)
		if c.regs.getFlag(FlagCarry) {
			carryIn = 0x80
		}
		c.regs.setFlag(FlagCarry, v&0x01 != 0)
		return v>>1 | carryIn
	})
	return 0
}
