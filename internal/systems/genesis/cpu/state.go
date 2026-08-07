package cpu

// Snapshot captures everything needed to resume execution exactly
// where it left off: every visible register plus the internal state
// gob would otherwise silently drop (unexported fields never survive
// encoding a struct that embeds them directly).
type Snapshot struct {
	D, A            [8]uint32
	PC              uint32
	SR              uint16
	USP, SSP        uint32
	Stopped         bool
	PendingIRQLevel byte
}

// Snapshot captures the CPU's full architectural state.
func (c *CPU) Snapshot() Snapshot {
	return Snapshot{
		D:               c.regs.D,
		A:               c.regs.A,
		PC:              c.regs.PC,
		SR:              c.regs.SR,
		USP:             c.regs.usp,
		SSP:             c.regs.ssp,
		Stopped:         c.stopped,
		PendingIRQLevel: c.pendingIRQLevel,
	}
}

// Restore reinstates a previously captured Snapshot.
func (c *CPU) Restore(s Snapshot) {
	c.regs.D = s.D
	c.regs.A = s.A
	c.regs.PC = s.PC
	c.regs.SR = s.SR
	c.regs.usp = s.USP
	c.regs.ssp = s.SSP
	c.stopped = s.Stopped
	c.pendingIRQLevel = s.PendingIRQLevel
}
