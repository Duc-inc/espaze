// Package cpu implements the Game Boy Advance's ARM7TDMI from
// scratch: both its 32-bit ARM and 16-bit Thumb instruction sets: This
// project has no BIOS (none is redistributable), so instead of the
// real boot ROM handing off to cartridge code, Reset starts execution
// directly at the cartridge entry point ($08000000) in ARM state -
// exactly where real hardware's BIOS would jump to after its own
// startup, so cartridge code itself is unaffected. THUMB mode gets
// the more complete implementation since it's what the large majority
// of commercial GBA game code actually runs in for code density; ARM
// mode covers the common instruction classes (data processing,
// branch, single/block data transfer, multiply) rather than every
// documented encoding.
package cpu

const cartridgeEntry = 0x08000000

// CPU is a from-scratch ARM7TDMI interpreter.
type CPU struct {
	regs registers
	bus  Bus

	pendingIRQ bool
}

// New wires a CPU to its bus and resets it.
func New(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset starts execution at the cartridge entry point in ARM state,
// supervisor mode, both interrupts masked - matching the CPU state
// real BIOS code would hand off in.
func (c *CPU) Reset() {
	c.regs = registers{}
	c.regs.CPSR = modeSupervisor | FlagIRQD | FlagFIQD
	c.regs.R[13] = 0x03007F00 // stack pointers real BIOS init code sets up
	c.regs.switchMode(modeIRQ)
	c.regs.R[13] = 0x03007FA0
	c.regs.switchMode(modeSupervisor)
	c.regs.R[15] = cartridgeEntry
	c.pendingIRQ = false
}

// PC exposes the program counter, mainly for tests.
func (c *CPU) PC() uint32 { return c.regs.R[15] }

// TriggerIRQ latches a hardware interrupt line.
func (c *CPU) TriggerIRQ() { c.pendingIRQ = true }

// Step services a pending IRQ (if unmasked), then executes exactly one
// instruction, returning an approximate cycle cost. Real ARM7TDMI
// timing depends on bus width/wait-states per access; this project
// uses a fixed, documented approximation instead of modeling that.
func (c *CPU) Step() int {
	if c.pendingIRQ && !c.regs.getFlag(FlagIRQD) {
		c.pendingIRQ = false
		return c.enterException(modeIRQ, 0x00000018, 4)
	}

	if c.regs.thumb() {
		return c.stepThumb()
	}
	return c.stepARM()
}

// enterException saves CPSR/PC into the target mode's SPSR/LR, masks
// IRQs, switches state to ARM, and jumps to the vector.
func (c *CPU) enterException(mode uint32, vector uint32, pcAdjust uint32) int {
	oldCPSR := c.regs.CPSR
	returnPC := c.regs.R[15] + pcAdjust

	c.regs.switchMode(mode)
	c.regs.spsr[modeIndex(mode)] = oldCPSR
	c.regs.R[14] = returnPC
	c.regs.setFlag(FlagThumb, false)
	c.regs.setFlag(FlagIRQD, true)
	c.regs.R[15] = vector
	return 3
}

func (c *CPU) currentSPSR() uint32 { return c.regs.spsr[modeIndex(c.regs.mode())] }
func (c *CPU) restoreFromSPSR()    { c.regs.CPSR = c.currentSPSR() }
