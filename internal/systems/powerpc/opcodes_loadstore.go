package powerpc

func (c *CPU) effectiveAddr(instr uint32) uint32 {
	rA := fieldRA(instr)
	return c.gprOrZero(rA) + uint32(fieldSimm(instr))
}

// effectiveAddrIndexed computes an X-form indexed effective address
// (rA + rB, with rA=0 meaning literal 0 like every other effective-
// address calculation) - the addressing mode lwzx/stwx and friends
// use, common for real compiled array-indexing code.
func (c *CPU) effectiveAddrIndexed(instr uint32) uint32 {
	return c.gprOrZero(fieldRA(instr)) + c.regs.GPR[fieldRB(instr)]
}

func init() {
	setPrimary(32, func(c *CPU, instr uint32) int { // lwz
		c.regs.GPR[fieldRD(instr)] = c.bus.Read32(c.effectiveAddr(instr))
		return 3
	})
	setPrimary(33, func(c *CPU, instr uint32) int { // lwzu
		addr := c.effectiveAddr(instr)
		c.regs.GPR[fieldRD(instr)] = c.bus.Read32(addr)
		c.regs.GPR[fieldRA(instr)] = addr
		return 3
	})
	setPrimary(34, func(c *CPU, instr uint32) int { // lbz
		c.regs.GPR[fieldRD(instr)] = uint32(c.bus.Read8(c.effectiveAddr(instr)))
		return 3
	})
	setPrimary(35, func(c *CPU, instr uint32) int { // lbzu
		addr := c.effectiveAddr(instr)
		c.regs.GPR[fieldRD(instr)] = uint32(c.bus.Read8(addr))
		c.regs.GPR[fieldRA(instr)] = addr
		return 3
	})
	setPrimary(40, func(c *CPU, instr uint32) int { // lhz
		c.regs.GPR[fieldRD(instr)] = uint32(c.bus.Read16(c.effectiveAddr(instr)))
		return 3
	})
	setPrimary(41, func(c *CPU, instr uint32) int { // lhzu
		addr := c.effectiveAddr(instr)
		c.regs.GPR[fieldRD(instr)] = uint32(c.bus.Read16(addr))
		c.regs.GPR[fieldRA(instr)] = addr
		return 3
	})

	setPrimary(36, func(c *CPU, instr uint32) int { // stw
		c.bus.Write32(c.effectiveAddr(instr), c.regs.GPR[fieldRD(instr)])
		return 3
	})
	setPrimary(37, func(c *CPU, instr uint32) int { // stwu
		addr := c.effectiveAddr(instr)
		c.bus.Write32(addr, c.regs.GPR[fieldRD(instr)])
		c.regs.GPR[fieldRA(instr)] = addr
		return 3
	})
	setPrimary(38, func(c *CPU, instr uint32) int { // stb
		c.bus.Write8(c.effectiveAddr(instr), byte(c.regs.GPR[fieldRD(instr)]))
		return 3
	})
	setPrimary(39, func(c *CPU, instr uint32) int { // stbu
		addr := c.effectiveAddr(instr)
		c.bus.Write8(addr, byte(c.regs.GPR[fieldRD(instr)]))
		c.regs.GPR[fieldRA(instr)] = addr
		return 3
	})
	setPrimary(44, func(c *CPU, instr uint32) int { // sth
		c.bus.Write16(c.effectiveAddr(instr), uint16(c.regs.GPR[fieldRD(instr)]))
		return 3
	})
	setPrimary(45, func(c *CPU, instr uint32) int { // sthu
		addr := c.effectiveAddr(instr)
		c.bus.Write16(addr, uint16(c.regs.GPR[fieldRD(instr)]))
		c.regs.GPR[fieldRA(instr)] = addr
		return 3
	})
	setPrimary(42, func(c *CPU, instr uint32) int { // lha (sign-extended halfword)
		c.regs.GPR[fieldRD(instr)] = uint32(int32(int16(c.bus.Read16(c.effectiveAddr(instr)))))
		return 3
	})
	setPrimary(43, func(c *CPU, instr uint32) int { // lhau
		addr := c.effectiveAddr(instr)
		c.regs.GPR[fieldRD(instr)] = uint32(int32(int16(c.bus.Read16(addr))))
		c.regs.GPR[fieldRA(instr)] = addr
		return 3
	})

	setExt31(23, func(c *CPU, instr uint32) int { // lwzx
		c.regs.GPR[fieldRD(instr)] = c.bus.Read32(c.effectiveAddrIndexed(instr))
		return 3
	})
	setExt31(87, func(c *CPU, instr uint32) int { // lbzx
		c.regs.GPR[fieldRD(instr)] = uint32(c.bus.Read8(c.effectiveAddrIndexed(instr)))
		return 3
	})
	setExt31(279, func(c *CPU, instr uint32) int { // lhzx
		c.regs.GPR[fieldRD(instr)] = uint32(c.bus.Read16(c.effectiveAddrIndexed(instr)))
		return 3
	})
	setExt31(343, func(c *CPU, instr uint32) int { // lhax
		c.regs.GPR[fieldRD(instr)] = uint32(int32(int16(c.bus.Read16(c.effectiveAddrIndexed(instr)))))
		return 3
	})
	setExt31(151, func(c *CPU, instr uint32) int { // stwx
		c.bus.Write32(c.effectiveAddrIndexed(instr), c.regs.GPR[fieldRD(instr)])
		return 3
	})
	setExt31(215, func(c *CPU, instr uint32) int { // stbx
		c.bus.Write8(c.effectiveAddrIndexed(instr), byte(c.regs.GPR[fieldRD(instr)]))
		return 3
	})
	setExt31(407, func(c *CPU, instr uint32) int { // sthx
		c.bus.Write16(c.effectiveAddrIndexed(instr), uint16(c.regs.GPR[fieldRD(instr)]))
		return 3
	})

	setPrimary(46, func(c *CPU, instr uint32) int { // lmw - load GPRs rD..r31 from consecutive words
		addr := c.effectiveAddr(instr)
		rD := fieldRD(instr)
		for r := rD; r <= 31; r++ {
			c.regs.GPR[r] = c.bus.Read32(addr)
			addr += 4
		}
		return 2 + int(32-rD)
	})
	setPrimary(47, func(c *CPU, instr uint32) int { // stmw - store GPRs rD..r31 to consecutive words
		addr := c.effectiveAddr(instr)
		rD := fieldRD(instr)
		for r := rD; r <= 31; r++ {
			c.bus.Write32(addr, c.regs.GPR[r])
			addr += 4
		}
		return 2 + int(32-rD)
	})
}
