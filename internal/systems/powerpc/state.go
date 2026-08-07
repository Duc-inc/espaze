package powerpc

// Snapshot captures the CPU's full register state.
type Snapshot struct {
	Regs registers
}

// Snapshot captures the CPU's current state.
func (c *CPU) Snapshot() Snapshot { return Snapshot{Regs: c.regs} }

// Restore reinstates a previously captured Snapshot.
func (c *CPU) Restore(s Snapshot) { c.regs = s.Regs }
