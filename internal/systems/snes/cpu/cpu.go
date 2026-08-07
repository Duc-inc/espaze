// Package cpu implements the SNES's 65816 from scratch: a 16-bit
// extension of the 6502/65C02 family (this project's NES/Atari 2600
// cores both implement the plain 6502) with independently switchable
// 8/16-bit accumulator and index registers, a 24-bit address space via
// Direct Page/Data Bank/Program Bank registers, and an "emulation
// mode" that makes it behave like a stock 6502 for backward-
// compatible boot code before games switch to native mode. Coverage
// here favors the instruction classes and addressing modes real SNES
// games actually rely on rather than the full documented matrix
// (some rarely-used addressing modes and decimal-mode arithmetic
// aren't implemented).
package cpu

const resetVector = 0x00FFFC

// CPU is a from-scratch 65816 interpreter.
type CPU struct {
	regs registers
	bus  Bus

	pendingNMI bool
	pendingIRQ bool
	stopped    bool
}

// New wires a CPU to its bus and resets it.
func New(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset reproduces the 65816's own reset sequence: it always starts
// in emulation mode (8-bit A/X/Y, stack pinned to page 1), regardless
// of what mode a previous run left it in.
func (c *CPU) Reset() {
	c.regs = registers{E: true, P: FlagIndex8 | FlagAccum8 | FlagIRQD}
	c.regs.S = 0x01FD
	c.regs.PC = c.read16(resetVector)
	c.pendingNMI = false
	c.pendingIRQ = false
	c.stopped = false
}

// PC exposes the 24-bit program address, mainly for tests.
func (c *CPU) PC() uint32 { return uint32(c.regs.PBR)<<16 | uint32(c.regs.PC) }

// TriggerNMI/TriggerIRQ latch the corresponding interrupt line.
func (c *CPU) TriggerNMI() { c.pendingNMI = true }
func (c *CPU) TriggerIRQ() { c.pendingIRQ = true }

func (c *CPU) read8(addr uint32) byte     { return c.bus.Read8(addr & 0xFFFFFF) }
func (c *CPU) write8(addr uint32, v byte) { c.bus.Write8(addr&0xFFFFFF, v) }

func (c *CPU) read16(addr uint32) uint16 {
	lo := uint16(c.read8(addr))
	hi := uint16(c.read8(addr + 1))
	return lo | hi<<8
}

func (c *CPU) write16(addr uint32, v uint16) {
	c.write8(addr, byte(v))
	c.write8(addr+1, byte(v>>8))
}

func (c *CPU) pcAddr() uint32 { return uint32(c.regs.PBR)<<16 | uint32(c.regs.PC) }

func (c *CPU) fetch8() byte {
	v := c.read8(c.pcAddr())
	c.regs.PC++
	return v
}

func (c *CPU) fetch16() uint16 {
	lo := uint16(c.fetch8())
	hi := uint16(c.fetch8())
	return lo | hi<<8
}

func (c *CPU) fetch24() uint32 {
	lo := uint32(c.fetch16())
	hi := uint32(c.fetch8())
	return lo | hi<<16
}

// push8/pop8 and push16/pop16 operate at whatever width the stack
// pointer's own emulation-mode pinning allows - in emulation mode S's
// high byte always reads back as 0x01 after any push/pop, matching
// real hardware.
func (c *CPU) push8(v byte) {
	c.write8(uint32(c.regs.S), v)
	c.regs.S--
	if c.regs.E {
		c.regs.S = c.regs.S&0x00FF | 0x0100
	}
}

func (c *CPU) pop8() byte {
	c.regs.S++
	if c.regs.E {
		c.regs.S = c.regs.S&0x00FF | 0x0100
	}
	return c.read8(uint32(c.regs.S))
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

// Step services a pending NMI/IRQ, then executes exactly one
// instruction, returning an approximate cycle cost - this project
// doesn't reproduce the real chip's per-addressing-mode/per-bank-
// crossing timing table.
func (c *CPU) Step() int {
	if c.pendingNMI {
		c.pendingNMI = false
		c.stopped = false
		return c.serviceInterrupt(c.vectorNMI())
	}
	if c.pendingIRQ && !c.regs.getFlag(FlagIRQD) {
		c.pendingIRQ = false
		c.stopped = false
		return c.serviceInterrupt(c.vectorIRQ())
	}
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

// vectorNMI/vectorIRQ pick the emulation- or native-mode vector -
// real hardware genuinely uses a different address depending on E.
func (c *CPU) vectorNMI() uint32 {
	if c.regs.E {
		return 0x00FFFA
	}
	return 0x00FFEA
}

func (c *CPU) vectorIRQ() uint32 {
	if c.regs.E {
		return 0x00FFFE
	}
	return 0x00FFEE
}

func (c *CPU) serviceInterrupt(vector uint32) int {
	if !c.regs.E {
		c.push8(c.regs.PBR)
	}
	c.push16(c.regs.PC)
	c.push8(c.regs.P)
	c.regs.setFlag(FlagIRQD, true)
	c.regs.setFlag(FlagDecimal, false)
	c.regs.PBR = 0
	c.regs.PC = c.read16(vector)
	return 8
}
