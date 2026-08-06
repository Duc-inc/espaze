package cpu

// shiftOps lists the 8 CB rotate/shift/swap operations in the order the
// 0x00-0x3F range encodes them.
var shiftOps = [8]func(c *CPU, v byte) byte{
	func(c *CPU, v byte) byte { return c.rlc(v) },
	func(c *CPU, v byte) byte { return c.rrc(v) },
	func(c *CPU, v byte) byte { return c.rl(v) },
	func(c *CPU, v byte) byte { return c.rr(v) },
	func(c *CPU, v byte) byte { return c.sla(v) },
	func(c *CPU, v byte) byte { return c.sra(v) },
	func(c *CPU, v byte) byte { return c.swap(v) },
	func(c *CPU, v byte) byte { return c.srl(v) },
}

func init() {
	for opIdx := byte(0); opIdx < 8; opIdx++ {
		op := shiftOps[opIdx]
		for r := byte(0); r < 8; r++ {
			opcode := opIdx*8 + r
			reg := r
			cycles := 8
			if reg == 6 {
				cycles = 16
			}
			cbTable[opcode] = func(c *CPU) int {
				c.writeR8(reg, op(c, c.readR8(reg)))
				return cycles
			}
		}
	}

	for bit := byte(0); bit < 8; bit++ {
		for r := byte(0); r < 8; r++ {
			b, reg := bit, r

			bitCycles := 8
			if reg == 6 {
				bitCycles = 12
			}
			cbTable[0x40+b*8+reg] = func(c *CPU) int {
				c.bit(c.readR8(reg), b)
				return bitCycles
			}

			rsCycles := 8
			if reg == 6 {
				rsCycles = 16
			}
			cbTable[0x80+b*8+reg] = func(c *CPU) int {
				c.writeR8(reg, c.readR8(reg)&^(1<<b))
				return rsCycles
			}
			cbTable[0xC0+b*8+reg] = func(c *CPU) int {
				c.writeR8(reg, c.readR8(reg)|(1<<b))
				return rsCycles
			}
		}
	}
}
