package cpu

// Snapshot captures everything needed to resume execution exactly
// where it left off.
type Snapshot struct {
	Regs       registers
	PendingNMI bool
	PendingIRQ bool
	Stopped    bool
}

// Snapshot captures the CPU's current state.
func (c *CPU) Snapshot() Snapshot {
	return Snapshot{Regs: c.regs, PendingNMI: c.pendingNMI, PendingIRQ: c.pendingIRQ, Stopped: c.stopped}
}

// Restore reinstates a previously captured Snapshot.
func (c *CPU) Restore(s Snapshot) {
	c.regs = s.Regs
	c.pendingNMI = s.PendingNMI
	c.pendingIRQ = s.PendingIRQ
	c.stopped = s.Stopped
}
