package cpu

// Snapshot captures everything needed to resume execution exactly
// where it left off: the full register file (including every banked
// mode's SP/LR/SPSR - gob silently drops unexported fields, so those
// are flattened into exported ones here rather than embedding
// registers directly) and any latched IRQ.
type Snapshot struct {
	R          [16]uint32
	CPSR       uint32
	BankedSP   [5]uint32
	BankedLR   [5]uint32
	SPSR       [5]uint32
	PendingIRQ bool
}

// Snapshot captures the CPU's current state.
func (c *CPU) Snapshot() Snapshot {
	return Snapshot{
		R: c.regs.R, CPSR: c.regs.CPSR,
		BankedSP: c.regs.bankedSP, BankedLR: c.regs.bankedLR, SPSR: c.regs.spsr,
		PendingIRQ: c.pendingIRQ,
	}
}

// Restore reinstates a previously captured Snapshot.
func (c *CPU) Restore(s Snapshot) {
	c.regs.R = s.R
	c.regs.CPSR = s.CPSR
	c.regs.bankedSP = s.BankedSP
	c.regs.bankedLR = s.BankedLR
	c.regs.spsr = s.SPSR
	c.pendingIRQ = s.PendingIRQ
}
