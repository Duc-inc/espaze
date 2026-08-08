package powerpc

func (c *CPU) gprOrZero(r uint32) uint32 {
	if r == 0 {
		return 0
	}
	return c.regs.GPR[r]
}

func init() {
	setPrimary(14, func(c *CPU, instr uint32) int { // addi
		rD, rA := fieldRD(instr), fieldRA(instr)
		c.regs.GPR[rD] = c.gprOrZero(rA) + uint32(fieldSimm(instr))
		return 2
	})
	setPrimary(15, func(c *CPU, instr uint32) int { // addis
		rD, rA := fieldRD(instr), fieldRA(instr)
		c.regs.GPR[rD] = c.gprOrZero(rA) + uint32(fieldSimm(instr))<<16
		return 2
	})
	setPrimary(7, func(c *CPU, instr uint32) int { // mulli
		rD, rA := fieldRD(instr), fieldRA(instr)
		c.regs.GPR[rD] = uint32(int32(c.regs.GPR[rA]) * fieldSimm(instr))
		return 3
	})

	setExt31(266, func(c *CPU, instr uint32) int { // add
		rD, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := c.regs.GPR[rA] + c.regs.GPR[rB]
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(40, func(c *CPU, instr uint32) int { // subf
		rD, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := c.regs.GPR[rB] - c.regs.GPR[rA]
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(235, func(c *CPU, instr uint32) int { // mullw
		rD, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := uint32(int32(c.regs.GPR[rA]) * int32(c.regs.GPR[rB]))
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 4
	})
	setExt31(491, func(c *CPU, instr uint32) int { // divw
		rD, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		if c.regs.GPR[rB] == 0 {
			return 4
		}
		result := uint32(int32(c.regs.GPR[rA]) / int32(c.regs.GPR[rB]))
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 20
	})
	setExt31(459, func(c *CPU, instr uint32) int { // divwu
		rD, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		if c.regs.GPR[rB] == 0 {
			return 4
		}
		result := c.regs.GPR[rA] / c.regs.GPR[rB]
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 20
	})
	setExt31(75, func(c *CPU, instr uint32) int { // mulhw
		rD, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		product := int64(int32(c.regs.GPR[rA])) * int64(int32(c.regs.GPR[rB]))
		result := uint32(product >> 32)
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 4
	})
	setExt31(11, func(c *CPU, instr uint32) int { // mulhwu
		rD, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		product := uint64(c.regs.GPR[rA]) * uint64(c.regs.GPR[rB])
		result := uint32(product >> 32)
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 4
	})
	setExt31(104, func(c *CPU, instr uint32) int { // neg
		rD, rA := fieldRD(instr), fieldRA(instr)
		result := -c.regs.GPR[rA]
		c.regs.GPR[rD] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})

	setPrimary(8, func(c *CPU, instr uint32) int { // subfic
		rD, rA := fieldRD(instr), fieldRA(instr)
		result, carry := addWithCarry(^c.regs.GPR[rA], uint32(fieldSimm(instr)), 1)
		c.regs.GPR[rD] = result
		c.regs.setXER(XERCarry, carry)
		return 2
	})
	setPrimary(12, func(c *CPU, instr uint32) int { // addic
		rD, rA := fieldRD(instr), fieldRA(instr)
		result, carry := addWithCarry(c.regs.GPR[rA], uint32(fieldSimm(instr)), 0)
		c.regs.GPR[rD] = result
		c.regs.setXER(XERCarry, carry)
		return 2
	})
	setPrimary(13, func(c *CPU, instr uint32) int { // addic.
		rD, rA := fieldRD(instr), fieldRA(instr)
		result, carry := addWithCarry(c.regs.GPR[rA], uint32(fieldSimm(instr)), 0)
		c.regs.GPR[rD] = result
		c.regs.setXER(XERCarry, carry)
		c.regs.setCR0(result)
		return 2
	})
}

// addWithCarry adds a, b, and a 1-bit carry-in, returning the 32-bit
// result and whether the 33-bit sum carried out - the primitive
// subfic/addic (and real hardware's addc/adde) build on. subfic uses
// it as SIMM - rA by passing ^rA with carry-in 1 (standard two's-
// complement subtraction via addition).
func addWithCarry(a, b, carryIn uint32) (result uint32, carryOut bool) {
	sum := uint64(a) + uint64(b) + uint64(carryIn)
	return uint32(sum), sum > 0xFFFFFFFF
}
