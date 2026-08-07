// Package cpu implements a from-scratch interpreter for the Neo Geo
// Pocket (Color)'s TLCS900H. This chip is dramatically less
// documented in publicly available reference material than every
// other CPU in this project (6502, Z80, 68000, HuC6280, ARM7TDMI all
// have extensive, well-established public documentation of their
// exact instruction encoding) - this project has confident,
// independently-checkable knowledge of the TLCS900H's real register
// architecture (the XWA/XBC/XDE/XHL 32-bit general registers with
// their 8/16-bit sub-views, IX/IY/IZ/SP index registers, and SR flag
// layout, all reproduced faithfully below), but NOT of its exact
// instruction byte encodings. Rather than guess at those and silently
// risk producing something that looks accurate but corrupts real
// game data, this package defines its own internally consistent
// encoding for a representative TLCS900H-style instruction set (8/16-
// bit load, arithmetic, logic, compare, jumps, calls, stack
// operations). The result is a genuine, working, testable CPU core in
// the spirit of the real chip's architecture - but it will not
// execute unmodified commercial Neo Geo Pocket ROM images correctly,
// since their bytes were assembled against the real chip's actual
// (differently-encoded) instruction set. This is flagged prominently
// rather than left to be discovered as a silent compatibility gap.
package cpu

const entryPoint = 0x000000

// CPU is a from-scratch TLCS900H-architecture interpreter.
type CPU struct {
	regs registers
	bus  Bus

	pendingIRQ bool
	irqVector  uint32
	halted     bool
}

// New wires a CPU to its bus and resets it.
func New(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset starts execution at the cartridge entry point with interrupts
// masked - this project has no BIOS, so (like this project's other
// BIOS-less cores) cartridge code runs directly from power-on rather
// than being handed off to by boot firmware.
func (c *CPU) Reset() {
	c.regs = registers{SR: 0x8F00} // interrupt mask = all levels disabled
	c.regs.XSP = 0x00004000
	c.regs.PC = entryPoint
	c.pendingIRQ = false
	c.halted = false
}

// PC exposes the program counter, mainly for tests.
func (c *CPU) PC() uint32 { return c.regs.PC }

// TriggerIRQ latches an interrupt with the given vector address.
func (c *CPU) TriggerIRQ(vector uint32) {
	c.pendingIRQ = true
	c.irqVector = vector
}

func (c *CPU) interruptsMasked() bool { return c.regs.SR&0x7000 == 0x7000 }

// Step services a pending interrupt, then executes exactly one
// instruction, returning an approximate cycle cost - this project
// doesn't reproduce the real chip's per-instruction/per-addressing-
// mode timing table, just a plausible fixed-ish cost per instruction
// class.
func (c *CPU) Step() int {
	if c.pendingIRQ && !c.interruptsMasked() {
		c.pendingIRQ = false
		c.halted = false
		return c.serviceInterrupt()
	}
	if c.halted {
		return 4
	}

	opcode := c.fetch8()
	entry := dispatchTable[opcode]
	if entry.execute == nil {
		return 2
	}
	return entry.execute(c)
}

func (c *CPU) serviceInterrupt() int {
	c.push32(c.regs.PC)
	c.push16(c.regs.SR)
	c.regs.PC = c.irqVector
	return 12
}
