// Package spc700 implements the SNES's audio coprocessor CPU from
// scratch. Like this project's Neo Geo Pocket TLCS900H core, the
// SPC700's exact instruction byte encoding isn't confidently known
// here (public documentation of it is far sparser than the 65816's),
// so this package defines its own internally consistent encoding for
// a representative SPC700-style instruction set (built on its real,
// well-documented register architecture: A/X/Y, an 8-bit SP always in
// page 1, the P flag's direct-page-base quirk, and PSW's flag layout)
// rather than guess at real byte values and risk silently producing
// something that looks accurate but isn't. It will not run unmodified
// commercial SNES sound-driver code correctly - flagged here rather
// than left as a silent gap.
package spc700

const entryPoint = 0xFFC0 // where this project's boot ROM stub, if any, would start

// CPU is a from-scratch SPC700-architecture interpreter.
type CPU struct {
	regs registers
	bus  Bus

	stopped bool
}

// New wires a CPU to its bus and resets it.
func New(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset starts execution at a fixed entry point with SP at the top of
// page 1 - this project has no IPL boot ROM (none is redistributable),
// so the main CPU is expected to upload driver code via the shared
// I/O ports and reset this CPU (via the port bridge) once ready.
func (c *CPU) Reset() {
	c.regs = registers{SP: 0xEF}
	c.regs.PC = entryPoint
	c.stopped = false
}

// PC exposes the program counter, mainly for tests.
func (c *CPU) PC() uint16 { return c.regs.PC }

func (c *CPU) read8(addr uint16) byte     { return c.bus.Read8(addr) }
func (c *CPU) write8(addr uint16, v byte) { c.bus.Write8(addr, v) }

func (c *CPU) fetch8() byte {
	v := c.read8(c.regs.PC)
	c.regs.PC++
	return v
}

func (c *CPU) fetch16() uint16 {
	lo := uint16(c.fetch8())
	hi := uint16(c.fetch8())
	return lo | hi<<8
}

func (c *CPU) push8(v byte) {
	c.write8(0x0100|uint16(c.regs.SP), v)
	c.regs.SP--
}

func (c *CPU) pop8() byte {
	c.regs.SP++
	return c.read8(0x0100 | uint16(c.regs.SP))
}

func (c *CPU) push16(v uint16) {
	c.push8(byte(v >> 8))
	c.push8(byte(v))
}

func (c *CPU) pop16() uint16 {
	lo := uint16(c.pop8())
	hi := uint16(c.pop8())
	return lo | hi<<8
}

// Step executes exactly one instruction, returning an approximate
// cycle cost.
func (c *CPU) Step() int {
	if c.stopped {
		return 2
	}
	opcode := c.fetch8()
	entry := dispatchTable[opcode]
	if entry.execute == nil {
		return 2
	}
	return entry.execute(c)
}
