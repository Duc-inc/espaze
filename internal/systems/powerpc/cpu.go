package powerpc

// CPU is a from-scratch 32-bit PowerPC interpreter.
type CPU struct {
	regs registers
	bus  Bus
}

// New wires a CPU to its bus and resets it.
func New(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset clears every register and starts execution at address 0 -
// real hardware's actual reset vector and boot ROM handoff aren't
// modeled, consistent with this package not being a playable system.
func (c *CPU) Reset() { c.regs = registers{} }

// PC exposes the program counter, mainly for tests.
func (c *CPU) PC() uint32 { return c.regs.PC }

func (c *CPU) fetch32() uint32 {
	v := c.bus.Read32(c.regs.PC)
	c.regs.PC += 4
	return v
}

// Step decodes and executes exactly one 32-bit instruction, returning
// an approximate cycle cost.
func (c *CPU) Step() int {
	op := c.fetch32()
	primary := op >> 26

	if fn, ok := primaryTable[primary]; ok {
		return fn(c, op)
	}
	return 2 // undefined: treated as a 2-cycle NOP
}
