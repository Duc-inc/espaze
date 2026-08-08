package powerpc

// CPU is a from-scratch 32-bit PowerPC interpreter.
type CPU struct {
	regs registers
	bus  Bus

	// SyscallHandler, if set, is called whenever the `sc` instruction
	// executes - real hardware raises a system-call exception instead;
	// this project has no exception vector table, so a direct hook is
	// this project's own stand-in for it, meant for high-level
	// emulation (HLE) of the small set of IPL/OS calls software
	// actually needs rather than interpreting real IPL firmware this
	// project doesn't have.
	SyscallHandler func(c *CPU)
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

// LR exposes the link register - the return address a "blr" (branch
// to link register) would jump to, which an HLE function call (see
// internal/systems/gamecube/hle) uses to return control to its caller
// without any real function body to run a real "blr" from.
func (c *CPU) LR() uint32 { return c.regs.LR }

// SetPC/SetGPR let a caller set up initial execution state - this
// project's own stand-in for what real IPL firmware would otherwise
// arrange before jumping into game code.
func (c *CPU) SetPC(addr uint32)        { c.regs.PC = addr }
func (c *CPU) SetGPR(reg int, v uint32) { c.regs.GPR[reg&0x1F] = v }
func (c *CPU) GPR(reg int) uint32       { return c.regs.GPR[reg&0x1F] }

func (c *CPU) fetch32() uint32 {
	v := c.read32(c.regs.PC)
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
