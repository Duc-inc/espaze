package cpu

// resetVector/nmiVector/irqVector are the fixed addresses the 6502 reads
// its program counter from when starting up or servicing an interrupt.
const (
	nmiVector   uint16 = 0xFFFA
	resetVector uint16 = 0xFFFC
	irqVector   uint16 = 0xFFFE
)

const stackBase uint16 = 0x0100

// CPU is a from-scratch implementation of the NES's 2A03: a stock 6502
// core (minus decimal mode, which the 2A03 wires off) plus whatever
// interrupt lines the PPU/APU drive.
type CPU struct {
	regs registers
	bus  Bus

	pendingNMI bool
	pendingIRQ bool
	halted     bool // set by an unimplemented/illegal opcode - not currently reachable, kept for cpu_test.go's coverage of the table
}

// New builds a CPU wired to bus. Reset still needs to be called before
// Step does anything meaningful (it loads PC from the reset vector).
func New(bus Bus) *CPU {
	return &CPU{bus: bus}
}

// Reset mimics the 6502's own reset sequence: SP moves back by 3 (as if
// three pushes happened without actually writing memory) and PC loads
// from the reset vector.
func (c *CPU) Reset() {
	c.regs.A, c.regs.X, c.regs.Y = 0, 0, 0
	c.regs.SP = 0xFD
	c.regs.P = FlagUnused | FlagInterrupt
	c.regs.PC = c.read16(resetVector)
	c.pendingNMI, c.pendingIRQ, c.halted = false, false, false
}

// PC exposes the program counter, mainly for tests.
func (c *CPU) PC() uint16 { return c.regs.PC }

// TriggerNMI/TriggerIRQ latch an interrupt line the PPU/APU can raise;
// it's serviced at the start of the next Step.
func (c *CPU) TriggerNMI() { c.pendingNMI = true }
func (c *CPU) TriggerIRQ() { c.pendingIRQ = true }

// Step executes exactly one instruction (after servicing any pending
// interrupt) and returns how many clock cycles it took.
func (c *CPU) Step() int {
	if c.pendingNMI {
		c.pendingNMI = false
		return c.serviceInterrupt(nmiVector, false) + 7
	}
	if c.pendingIRQ && !c.regs.getFlag(FlagInterrupt) {
		c.pendingIRQ = false
		return c.serviceInterrupt(irqVector, false) + 7
	}
	if c.halted {
		return 1
	}

	opcode := c.fetchByte()
	entry := opcodeTable[opcode]

	addr, pageCrossed := c.resolveOperand(entry.mode)
	cycles := entry.cycles
	if pageCrossed && entry.extraOnPageCross {
		cycles++
	}
	cycles += entry.execute(c, entry.mode, addr, pageCrossed)
	return cycles
}

func (c *CPU) serviceInterrupt(vector uint16, brk bool) int {
	c.push16(c.regs.PC)
	flags := c.regs.P | FlagUnused
	if brk {
		flags |= FlagBreak
	} else {
		flags &^= FlagBreak
	}
	c.push(flags)
	c.regs.setFlag(FlagInterrupt, true)
	c.regs.PC = c.read16(vector)
	return 0
}

func (c *CPU) fetchByte() byte {
	v := c.bus.Read(c.regs.PC)
	c.regs.PC++
	return v
}

func (c *CPU) fetch16() uint16 {
	lo := uint16(c.fetchByte())
	hi := uint16(c.fetchByte())
	return hi<<8 | lo
}

// read16 reproduces the 6502's zero-page/vector wraparound bug: the high
// byte is fetched from (addr+1) truncated to stay within the same page.
func (c *CPU) read16(addr uint16) uint16 {
	lo := uint16(c.bus.Read(addr))
	hiAddr := (addr & 0xFF00) | uint16(byte(addr)+1)
	hi := uint16(c.bus.Read(hiAddr))
	return hi<<8 | lo
}

func (c *CPU) push(v byte) {
	c.bus.Write(stackBase+uint16(c.regs.SP), v)
	c.regs.SP--
}

func (c *CPU) pop() byte {
	c.regs.SP++
	return c.bus.Read(stackBase + uint16(c.regs.SP))
}

func (c *CPU) push16(v uint16) {
	c.push(byte(v >> 8))
	c.push(byte(v))
}

func (c *CPU) pop16() uint16 {
	lo := uint16(c.pop())
	hi := uint16(c.pop())
	return hi<<8 | lo
}
