package cpu

// Snapshot captures every piece of CPU state needed to resume execution
// later: registers, the interrupt master enable machinery, and whether
// the CPU is halted/stopped.
type Snapshot struct {
	Regs    registers
	IME     bool
	EIDelay int
	Halted  bool
	Stopped bool
}

func (c *CPU) Snapshot() Snapshot {
	return Snapshot{
		Regs:    c.regs,
		IME:     c.ime,
		EIDelay: c.eiDelay,
		Halted:  c.halted,
		Stopped: c.stopped,
	}
}

func (c *CPU) Restore(s Snapshot) {
	c.regs = s.Regs
	c.ime = s.IME
	c.eiDelay = s.EIDelay
	c.halted = s.Halted
	c.stopped = s.Stopped
}
