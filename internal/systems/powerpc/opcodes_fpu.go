package powerpc

import "math"

func (c *CPU) read64(addr uint32) uint64 {
	return uint64(c.read32(addr))<<32 | uint64(c.read32(addr+4))
}

func (c *CPU) write64(addr uint32, v uint64) {
	c.write32(addr, uint32(v>>32))
	c.write32(addr+4, uint32(v))
}

func init() {
	setPrimary(50, func(c *CPU, instr uint32) int { // lfd
		c.regs.FPR[fieldRD(instr)] = math.Float64frombits(c.read64(c.effectiveAddr(instr)))
		return 4
	})
	setPrimary(51, func(c *CPU, instr uint32) int { // lfdu
		addr := c.effectiveAddr(instr)
		c.regs.FPR[fieldRD(instr)] = math.Float64frombits(c.read64(addr))
		c.regs.GPR[fieldRA(instr)] = addr
		return 4
	})
	setPrimary(48, func(c *CPU, instr uint32) int { // lfs
		bits := c.read32(c.effectiveAddr(instr))
		c.regs.FPR[fieldRD(instr)] = float64(math.Float32frombits(bits))
		return 4
	})
	setPrimary(49, func(c *CPU, instr uint32) int { // lfsu
		addr := c.effectiveAddr(instr)
		c.regs.FPR[fieldRD(instr)] = float64(math.Float32frombits(c.read32(addr)))
		c.regs.GPR[fieldRA(instr)] = addr
		return 4
	})

	setPrimary(54, func(c *CPU, instr uint32) int { // stfd
		c.write64(c.effectiveAddr(instr), math.Float64bits(c.regs.FPR[fieldRD(instr)]))
		return 4
	})
	setPrimary(55, func(c *CPU, instr uint32) int { // stfdu
		addr := c.effectiveAddr(instr)
		c.write64(addr, math.Float64bits(c.regs.FPR[fieldRD(instr)]))
		c.regs.GPR[fieldRA(instr)] = addr
		return 4
	})
	setPrimary(52, func(c *CPU, instr uint32) int { // stfs
		bits := math.Float32bits(float32(c.regs.FPR[fieldRD(instr)]))
		c.write32(c.effectiveAddr(instr), bits)
		return 4
	})
	setPrimary(53, func(c *CPU, instr uint32) int { // stfsu
		addr := c.effectiveAddr(instr)
		c.write32(addr, math.Float32bits(float32(c.regs.FPR[fieldRD(instr)])))
		c.regs.GPR[fieldRA(instr)] = addr
		return 4
	})

	setExt63A(21, func(c *CPU, instr uint32) int { // fadd
		frD, frA, frB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		c.regs.FPR[frD] = c.regs.FPR[frA] + c.regs.FPR[frB]
		return 4
	})
	setExt63A(20, func(c *CPU, instr uint32) int { // fsub
		frD, frA, frB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		c.regs.FPR[frD] = c.regs.FPR[frA] - c.regs.FPR[frB]
		return 4
	})
	setExt63A(25, func(c *CPU, instr uint32) int { // fmul (uses frC, not frB)
		frD, frA, frC := fieldRD(instr), fieldRA(instr), fieldFRC(instr)
		c.regs.FPR[frD] = c.regs.FPR[frA] * c.regs.FPR[frC]
		return 4
	})
	setExt63A(18, func(c *CPU, instr uint32) int { // fdiv
		frD, frA, frB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		c.regs.FPR[frD] = c.regs.FPR[frA] / c.regs.FPR[frB]
		return 20
	})

	setExt63(72, func(c *CPU, instr uint32) int { // fmr
		c.regs.FPR[fieldRD(instr)] = c.regs.FPR[fieldRB(instr)]
		return 4
	})
	setExt63(40, func(c *CPU, instr uint32) int { // fneg
		c.regs.FPR[fieldRD(instr)] = -c.regs.FPR[fieldRB(instr)]
		return 4
	})
	setExt63(264, func(c *CPU, instr uint32) int { // fabs
		c.regs.FPR[fieldRD(instr)] = math.Abs(c.regs.FPR[fieldRB(instr)])
		return 4
	})
	// Primary 59: single-precision counterparts of the above (same xo
	// values under a different primary opcode, standard PowerPC ISA
	// layout) - the result is rounded to float32 precision after the
	// double-precision computation, matching real hardware's own
	// single-precision arithmetic semantics.
	setExt59A(21, func(c *CPU, instr uint32) int { // fadds
		frD, frA, frB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		c.regs.FPR[frD] = float64(float32(c.regs.FPR[frA] + c.regs.FPR[frB]))
		return 4
	})
	setExt59A(20, func(c *CPU, instr uint32) int { // fsubs
		frD, frA, frB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		c.regs.FPR[frD] = float64(float32(c.regs.FPR[frA] - c.regs.FPR[frB]))
		return 4
	})
	setExt59A(25, func(c *CPU, instr uint32) int { // fmuls (uses frC, not frB)
		frD, frA, frC := fieldRD(instr), fieldRA(instr), fieldFRC(instr)
		c.regs.FPR[frD] = float64(float32(c.regs.FPR[frA] * c.regs.FPR[frC]))
		return 4
	})
	setExt59A(18, func(c *CPU, instr uint32) int { // fdivs
		frD, frA, frB := fieldRD(instr), fieldRA(instr), fieldRB(instr)
		c.regs.FPR[frD] = float64(float32(c.regs.FPR[frA] / c.regs.FPR[frB]))
		return 20
	})

	// Fused multiply-add family (A-form, frA*frC +/- frB, optionally
	// negated) - standard PowerPC ISA instructions computed as one
	// rounding step on real hardware; this project computes them as two
	// separate Go float64 operations, an acceptable simplification for
	// the tiny last-bit rounding difference that introduces.
	setExt63A(29, func(c *CPU, instr uint32) int { // fmadd
		frD, frA, frB, frC := fieldRD(instr), fieldRA(instr), fieldRB(instr), fieldFRC(instr)
		c.regs.FPR[frD] = c.regs.FPR[frA]*c.regs.FPR[frC] + c.regs.FPR[frB]
		return 4
	})
	setExt63A(28, func(c *CPU, instr uint32) int { // fmsub
		frD, frA, frB, frC := fieldRD(instr), fieldRA(instr), fieldRB(instr), fieldFRC(instr)
		c.regs.FPR[frD] = c.regs.FPR[frA]*c.regs.FPR[frC] - c.regs.FPR[frB]
		return 4
	})
	setExt63A(31, func(c *CPU, instr uint32) int { // fnmadd
		frD, frA, frB, frC := fieldRD(instr), fieldRA(instr), fieldRB(instr), fieldFRC(instr)
		c.regs.FPR[frD] = -(c.regs.FPR[frA]*c.regs.FPR[frC] + c.regs.FPR[frB])
		return 4
	})
	setExt63A(30, func(c *CPU, instr uint32) int { // fnmsub
		frD, frA, frB, frC := fieldRD(instr), fieldRA(instr), fieldRB(instr), fieldFRC(instr)
		c.regs.FPR[frD] = -(c.regs.FPR[frA]*c.regs.FPR[frC] - c.regs.FPR[frB])
		return 4
	})
	// Primary 59: single-precision counterparts, same rounding pattern
	// as fadds/fsubs/fmuls/fdivs above.
	setExt59A(29, func(c *CPU, instr uint32) int { // fmadds
		frD, frA, frB, frC := fieldRD(instr), fieldRA(instr), fieldRB(instr), fieldFRC(instr)
		c.regs.FPR[frD] = float64(float32(c.regs.FPR[frA]*c.regs.FPR[frC] + c.regs.FPR[frB]))
		return 4
	})
	setExt59A(28, func(c *CPU, instr uint32) int { // fmsubs
		frD, frA, frB, frC := fieldRD(instr), fieldRA(instr), fieldRB(instr), fieldFRC(instr)
		c.regs.FPR[frD] = float64(float32(c.regs.FPR[frA]*c.regs.FPR[frC] - c.regs.FPR[frB]))
		return 4
	})
	setExt59A(31, func(c *CPU, instr uint32) int { // fnmadds
		frD, frA, frB, frC := fieldRD(instr), fieldRA(instr), fieldRB(instr), fieldFRC(instr)
		c.regs.FPR[frD] = float64(float32(-(c.regs.FPR[frA]*c.regs.FPR[frC] + c.regs.FPR[frB])))
		return 4
	})
	setExt59A(30, func(c *CPU, instr uint32) int { // fnmsubs
		frD, frA, frB, frC := fieldRD(instr), fieldRA(instr), fieldRB(instr), fieldFRC(instr)
		c.regs.FPR[frD] = float64(float32(-(c.regs.FPR[frA]*c.regs.FPR[frC] - c.regs.FPR[frB])))
		return 4
	})

	setExt63(12, func(c *CPU, instr uint32) int { // frsp: round double to single precision
		c.regs.FPR[fieldRD(instr)] = float64(float32(c.regs.FPR[fieldRB(instr)]))
		return 4
	})
	setExt63A(23, func(c *CPU, instr uint32) int { // fsel frD,frA,frC,frB: (frA>=0.0) ? frC : frB
		frD, frA, frB, frC := fieldRD(instr), fieldRA(instr), fieldRB(instr), fieldFRC(instr)
		if c.regs.FPR[frA] >= 0 {
			c.regs.FPR[frD] = c.regs.FPR[frC]
		} else {
			c.regs.FPR[frD] = c.regs.FPR[frB]
		}
		return 4
	})
	setExt63(15, func(c *CPU, instr uint32) int { // fctiwz: convert to integer, truncate toward zero
		// Real hardware packs the result into the FPR's low 32 bits with
		// an implementation-specific high-word fill this project doesn't
		// claim to reproduce exactly; the low word (what a real stfd+lwz
		// pair, the standard way compiled code reads this result, would
		// see) is correct.
		v := c.regs.FPR[fieldRB(instr)]
		var i int32
		switch {
		case v >= 2147483647:
			i = 2147483647
		case v <= -2147483648:
			i = -2147483648
		default:
			i = int32(v)
		}
		c.regs.FPR[fieldRD(instr)] = math.Float64frombits(uint64(uint32(i)))
		return 4
	})

	setExt63(0, func(c *CPU, instr uint32) int { // fcmpu
		a, b := c.regs.FPR[fieldRA(instr)], c.regs.FPR[fieldRB(instr)]
		var field uint32
		switch {
		case a < b:
			field = 0x8
		case a > b:
			field = 0x4
		default:
			field = 0x2
		}
		c.regs.CR = c.regs.CR&0x0FFFFFFF | field<<28
		return 4
	})
}
