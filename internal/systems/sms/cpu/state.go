package cpu

// Snapshot captures everything needed to resume execution exactly
// where it left off.
type Snapshot struct {
	Regs       registers
	Halted     bool
	PendingNMI bool
	PendingINT bool
	IntData    byte
	EIDelay    bool
}

func (c *CPU) Snapshot() Snapshot {
	return Snapshot{
		Regs: c.regs, Halted: c.halted,
		PendingNMI: c.pendingNMI, PendingINT: c.pendingINT, IntData: c.intData,
		EIDelay: c.eiDelay,
	}
}

func (c *CPU) Restore(s Snapshot) {
	c.regs = s.Regs
	c.halted = s.Halted
	c.pendingNMI, c.pendingINT, c.intData = s.PendingNMI, s.PendingINT, s.IntData
	c.eiDelay = s.EIDelay
}
