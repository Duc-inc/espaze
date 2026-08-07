package spc700

// Snapshot captures everything needed to resume execution exactly
// where it left off.
type Snapshot struct {
	Regs    registers
	Stopped bool
}

// Snapshot captures the CPU's current state.
func (c *CPU) Snapshot() Snapshot { return Snapshot{Regs: c.regs, Stopped: c.stopped} }

// Restore reinstates a previously captured Snapshot.
func (c *CPU) Restore(s Snapshot) {
	c.regs = s.Regs
	c.stopped = s.Stopped
}
