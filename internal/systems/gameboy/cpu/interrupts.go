package cpu

// Interrupt vectors, in priority order (lowest bit = highest priority).
var interruptVectors = [5]uint16{0x40, 0x48, 0x50, 0x58, 0x60}

// checkInterrupts services the highest-priority pending, enabled
// interrupt (IF & IE), waking a halted CPU even if IME is off. Returns
// the T-cycles spent dispatching it, or 0 if nothing was serviced.
func (c *CPU) checkInterrupts() int {
	ifReg := c.bus.Read(0xFF0F)
	ieReg := c.bus.Read(0xFFFF)
	pending := ifReg & ieReg & 0x1F
	if pending == 0 {
		return 0
	}

	if c.halted {
		c.halted = false
	}
	if !c.ime {
		return 0
	}

	for i := 0; i < 5; i++ {
		bit := byte(1) << uint(i)
		if pending&bit == 0 {
			continue
		}
		c.ime = false
		c.bus.Write(0xFF0F, ifReg&^bit)
		c.push16(c.regs.PC)
		c.regs.PC = interruptVectors[i]
		return 20
	}
	return 0
}
