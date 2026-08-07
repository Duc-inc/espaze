package powerpc

func (c *CPU) effectiveAddr(instr uint32) uint32 {
	rA := fieldRA(instr)
	return c.gprOrZero(rA) + uint32(fieldSimm(instr))
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
}
