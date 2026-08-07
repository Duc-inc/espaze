package cpu

// Operand sizes, using the same 2-bit encoding several instruction
// families (MOVE, and the 00/01/10 size field on many others) use
// natively, so decoding rarely needs to translate between them.
const (
	sizeByte byte = 0
	sizeWord byte = 1
	sizeLong byte = 2
)

// Exception vector numbers this project actually raises.
const (
	vectorReset      = 0
	vectorIllegal    = 4
	vectorZeroDivide = 5
	vectorPrivilege  = 8
	vectorTrap0      = 32 // TRAP #0 through #15 follow sequentially
)

// CPU is a from-scratch Motorola 68000 interpreter. Its instruction set
// is decoded via a 65536-entry dispatch table (see dispatch.go), built
// once by matching each opcode's documented bit pattern - the standard
// approach every serious 68k emulator uses, since the encoding isn't
// regular enough to decode live the way the Z80/6502 cores in this
// project do.
type CPU struct {
	regs registers
	bus  Bus

	stopped bool

	pendingIRQLevel byte // 0 = none pending
}

// New wires a CPU to its memory bus and resets it (real 68000 hardware
// has no separate reset-then-run step: the reset exception itself loads
// the initial SSP and PC from the vector table at $000000/$000004).
func New(bus Bus) *CPU {
	ensureDispatchTable()
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset reproduces the 68000's own reset sequence: supervisor mode, all
// interrupts masked, SSP and PC loaded from the reset vector.
func (c *CPU) Reset() {
	c.regs = registers{SR: srSupervisor | srIPMask}
	c.regs.A[7] = c.bus.Read32(0)
	c.regs.ssp = c.regs.A[7]
	c.regs.PC = c.bus.Read32(4)
	c.stopped = false
	c.pendingIRQLevel = 0
}

// PC exposes the program counter, mainly for tests.
func (c *CPU) PC() uint32 { return c.regs.PC }

// TriggerIRQ latches an interrupt at the given priority (1-7); it's
// serviced once the CPU's own interrupt mask allows it through, exactly
// like a real autovectored interrupt line.
func (c *CPU) TriggerIRQ(level byte) {
	if level > c.pendingIRQLevel {
		c.pendingIRQLevel = level
	}
}

func (c *CPU) fetchWord() uint16 {
	v := c.bus.Read16(c.regs.PC)
	c.regs.PC += 2
	return v
}

func (c *CPU) fetchLong() uint32 {
	v := c.bus.Read32(c.regs.PC)
	c.regs.PC += 4
	return v
}

func (c *CPU) push32(v uint32) {
	c.regs.A[7] -= 4
	c.bus.Write32(c.regs.A[7], v)
}

func (c *CPU) pop32() uint32 {
	v := c.bus.Read32(c.regs.A[7])
	c.regs.A[7] += 4
	return v
}

// Step services any pending interrupt, then executes exactly one
// instruction, returning how many clock cycles it took.
func (c *CPU) Step() int {
	if c.pendingIRQLevel > 0 && c.pendingIRQLevel > c.regs.interruptMask() {
		return c.serviceInterrupt()
	}
	if c.stopped {
		return 4
	}

	opcode := c.fetchWord()
	entry := dispatchTable[opcode]
	if entry.execute == nil {
		return c.raiseException(vectorIllegal)
	}
	return entry.execute(c, opcode)
}

func (c *CPU) serviceInterrupt() int {
	level := c.pendingIRQLevel
	c.pendingIRQLevel = 0
	c.stopped = false

	oldSR := c.regs.SR
	c.regs.enterSupervisor()
	c.push32(c.regs.PC)
	c.push16FromSR(oldSR)
	c.regs.SR = (c.regs.SR &^ srIPMask) | uint16(level)<<8

	vector := uint32(24+level) * 4 // autovector 25-31 for levels 1-7
	c.regs.PC = c.bus.Read32(vector)
	return 44
}

func (c *CPU) push16FromSR(sr uint16) {
	c.regs.A[7] -= 2
	c.bus.Write16(c.regs.A[7], sr)
}

// raiseException implements the small set of synchronous traps this
// project actually needs: illegal instructions, division by zero,
// privilege violations, and the TRAP #n instruction.
func (c *CPU) raiseException(vector uint32) int {
	oldSR := c.regs.SR
	oldPC := c.regs.PC
	c.regs.enterSupervisor()
	c.push32(oldPC)
	c.push16FromSR(oldSR)
	c.regs.PC = c.bus.Read32(vector * 4)
	return 34
}
