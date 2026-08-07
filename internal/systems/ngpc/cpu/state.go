package cpu

// Snapshot captures everything needed to resume execution exactly
// where it left off.
type Snapshot struct {
	Regs       registers
	PendingIRQ bool
	IRQVector  uint32
	Halted     bool
}

// Snapshot captures the CPU's current state.
func (c *CPU) Snapshot() Snapshot {
	return Snapshot{Regs: c.regs, PendingIRQ: c.pendingIRQ, IRQVector: c.irqVector, Halted: c.halted}
}

// Restore reinstates a previously captured Snapshot.
func (c *CPU) Restore(s Snapshot) {
	c.regs = s.Regs
	c.pendingIRQ = s.PendingIRQ
	c.irqVector = s.IRQVector
	c.halted = s.Halted
}
