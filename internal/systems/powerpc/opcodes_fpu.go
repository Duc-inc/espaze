package powerpc

import "math"

func (c *CPU) read64(addr uint32) uint64 {
	return uint64(c.bus.Read32(addr))<<32 | uint64(c.bus.Read32(addr+4))
}

func (c *CPU) write64(addr uint32, v uint64) {
	c.bus.Write32(addr, uint32(v>>32))
	c.bus.Write32(addr+4, uint32(v))
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
		bits := c.bus.Read32(c.effectiveAddr(instr))
		c.regs.FPR[fieldRD(instr)] = float64(math.Float32frombits(bits))
		return 4
	})
	setPrimary(49, func(c *CPU, instr uint32) int { // lfsu
		addr := c.effectiveAddr(instr)
		c.regs.FPR[fieldRD(instr)] = float64(math.Float32frombits(c.bus.Read32(addr)))
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
		c.bus.Write32(c.effectiveAddr(instr), bits)
		return 4
	})
	setPrimary(53, func(c *CPU, instr uint32) int { // stfsu
		addr := c.effectiveAddr(instr)
		c.bus.Write32(addr, math.Float32bits(float32(c.regs.FPR[fieldRD(instr)])))
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
