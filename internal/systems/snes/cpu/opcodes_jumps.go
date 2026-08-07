package cpu

func init() {
	setOp(0x4C, func(c *CPU) int { c.regs.PC = c.fetch16(); return 3 })
	setOp(0x6C, func(c *CPU) int {
		ptr := uint32(c.fetch16())
		c.regs.PC = c.read16(ptr)
		return 5
	})
	setOp(0x7C, func(c *CPU) int {
		base := uint32(c.regs.PBR)<<16 | uint32(c.fetch16())
		ptr := (base + uint32(c.regs.X)) & 0xFFFFFF
		c.regs.PC = c.read16(ptr)
		return 6
	})
	setOp(0x5C, func(c *CPU) int {
		addr := c.fetch24()
		c.regs.PBR = byte(addr >> 16)
		c.regs.PC = uint16(addr)
		return 4
	})
	setOp(0xDC, func(c *CPU) int {
		ptr := uint32(c.fetch16())
		lo := uint32(c.read16(ptr))
		hi := uint32(c.read8(ptr + 2))
		c.regs.PBR = byte(hi)
		c.regs.PC = uint16(lo)
		return 6
	})

	setOp(0x20, func(c *CPU) int {
		target := c.fetch16()
		c.push16(c.regs.PC - 1)
		c.regs.PC = target
		return 6
	})
	setOp(0xFC, func(c *CPU) int {
		base := uint32(c.regs.PBR)<<16 | uint32(c.fetch16())
		c.push16(c.regs.PC - 1)
		ptr := (base + uint32(c.regs.X)) & 0xFFFFFF
		c.regs.PC = c.read16(ptr)
		return 8
	})
	setOp(0x22, func(c *CPU) int {
		addr := c.fetch24()
		c.push8(c.regs.PBR)
		c.push16(c.regs.PC - 1)
		c.regs.PBR = byte(addr >> 16)
		c.regs.PC = uint16(addr)
		return 8
	})

	setOp(0x60, func(c *CPU) int { c.regs.PC = c.pop16() + 1; return 6 })
	setOp(0x6B, func(c *CPU) int {
		c.regs.PC = c.pop16() + 1
		c.regs.PBR = c.pop8()
		return 6
	})
	setOp(0x40, func(c *CPU) int {
		c.regs.P = c.pop8()
		c.regs.PC = c.pop16()
		if !c.regs.E {
			c.regs.PBR = c.pop8()
		}
		return 7
	})

	setOp(0x00, func(c *CPU) int {
		c.regs.PC++ // BRK's signature byte is skipped on return
		vector := uint32(0x00FFE6)
		if c.regs.E {
			vector = 0x00FFFE
		}
		return c.serviceInterrupt(vector)
	})
	setOp(0x02, func(c *CPU) int {
		c.regs.PC++
		vector := uint32(0x00FFE4)
		if c.regs.E {
			vector = 0x00FFF4
		}
		return c.serviceInterrupt(vector)
	})
}
