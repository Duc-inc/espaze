package powerpc

// rotl32 rotates x left by n bits (0-31) - the primitive every M-form
// rotate/mask instruction (rlwinm/rlwimi/rlwnm) builds on. Go's shift
// operators already define x>>32 as 0 for a 32-bit value, so n==0
// needs no special case: rotl32(x, 0) == x<<0 | x>>32 == x.
func rotl32(x, n uint32) uint32 {
	return x<<n | x>>(32-n)
}

// rangeMask builds a standard (LSB-numbered) 32-bit mask covering
// PowerPC's own MB..ME bit range (both given in IBM's MSB-first
// numbering, 0-31), matching the well-documented rule real rotate/
// mask instructions use: if mb <= me, a contiguous run of 1-bits from
// mb to me; otherwise the range wraps around bit 31 back to bit 0.
func rangeMask(mb, me uint32) uint32 {
	if mb <= me {
		return bitRange(mb, me)
	}
	return bitRange(mb, 31) | bitRange(0, me)
}

// bitRange returns a mask of 1-bits for the inclusive IBM-numbered
// range [start, end], with no wraparound.
func bitRange(start, end uint32) uint32 {
	n := end - start + 1
	if n == 32 {
		return 0xFFFFFFFF
	}
	return (uint32(1)<<n - 1) << (31 - end)
}

func init() {
	setPrimary(21, func(c *CPU, instr uint32) int { // rlwinm
		rS, rA := fieldRD(instr), fieldRA(instr)
		mask := rangeMask(fieldMB(instr), fieldME(instr))
		result := rotl32(c.regs.GPR[rS], fieldSH(instr)) & mask
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setPrimary(20, func(c *CPU, instr uint32) int { // rlwimi
		rS, rA := fieldRD(instr), fieldRA(instr)
		mask := rangeMask(fieldMB(instr), fieldME(instr))
		rotated := rotl32(c.regs.GPR[rS], fieldSH(instr))
		result := (rotated & mask) | (c.regs.GPR[rA] & ^mask)
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setPrimary(23, func(c *CPU, instr uint32) int { // rlwnm
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		mask := rangeMask(fieldMB(instr), fieldME(instr))
		shift := c.regs.GPR[rB] & 0x1F
		result := rotl32(c.regs.GPR[rS], shift) & mask
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})

	setExt31(24, func(c *CPU, instr uint32) int { // slw
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := c.regs.GPR[rS] << (c.regs.GPR[rB] & 0x3F)
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(536, func(c *CPU, instr uint32) int { // srw
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result := c.regs.GPR[rS] >> (c.regs.GPR[rB] & 0x3F)
		c.regs.GPR[rA] = result
		if fieldRC(instr) {
			c.regs.setCR0(result)
		}
		return 2
	})
	setExt31(792, func(c *CPU, instr uint32) int { // sraw
		rS, rA, rB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		result, carry := arithShiftRight(int32(c.regs.GPR[rS]), c.regs.GPR[rB]&0x3F)
		c.regs.GPR[rA] = uint32(result)
		c.regs.setXER(XERCarry, carry)
		if fieldRC(instr) {
			c.regs.setCR0(uint32(result))
		}
		return 2
	})
	setExt31(824, func(c *CPU, instr uint32) int { // srawi
		rS, rA := fieldRD(instr), fieldRA(instr)
		result, carry := arithShiftRight(int32(c.regs.GPR[rS]), fieldSH(instr))
		c.regs.GPR[rA] = uint32(result)
		c.regs.setXER(XERCarry, carry)
		if fieldRC(instr) {
			c.regs.setCR0(uint32(result))
		}
		return 2
	})
}

// arithShiftRight implements sraw/srawi's real semantics: an
// arithmetic (sign-extending) right shift, plus XER's carry bit set
// whenever the source was negative and any 1-bit was shifted out -
// a real, well-documented (if easy to miss) quirk of these two
// instructions specifically, not shared by any other PowerPC shift.
func arithShiftRight(s int32, n uint32) (result int32, carry bool) {
	if n >= 32 {
		if s < 0 {
			return -1, true
		}
		return 0, false
	}
	result = s >> n
	if s < 0 && uint32(s)&(1<<n-1) != 0 {
		carry = true
	}
	return result, carry
}
