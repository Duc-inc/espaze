package cpu

// Snapshot captures everything needed to resume execution exactly
// where it left off: registers, the MMU's page table, the built-in
// timer/interrupt controller, and any latched NMI.
type Snapshot struct {
	Regs       registers
	MPR        [8]byte
	IRQ        irqSnapshot
	PendingNMI bool
}

type irqSnapshot struct {
	MaskIRQ2, MaskIRQ1, MaskTimer          bool
	PendingIRQ2, PendingIRQ1, PendingTimer bool
	TimerReload                            byte
	TimerCounter, TimerAccum               int
	TimerRunning                           bool
}

// Snapshot captures the CPU's current state.
func (c *CPU) Snapshot() Snapshot {
	return Snapshot{
		Regs: c.regs,
		MPR:  c.mmu.mpr,
		IRQ: irqSnapshot{
			MaskIRQ2: c.irq.maskIRQ2, MaskIRQ1: c.irq.maskIRQ1, MaskTimer: c.irq.maskTimer,
			PendingIRQ2: c.irq.pendingIRQ2, PendingIRQ1: c.irq.pendingIRQ1, PendingTimer: c.irq.pendingTimer,
			TimerReload: c.irq.timerReload, TimerCounter: c.irq.timerCounter,
			TimerAccum: c.irq.timerAccum, TimerRunning: c.irq.timerRunning,
		},
		PendingNMI: c.pendingNMI,
	}
}

// Restore reinstates a previously captured Snapshot.
func (c *CPU) Restore(s Snapshot) {
	c.regs = s.Regs
	c.mmu.mpr = s.MPR
	c.irq.maskIRQ2, c.irq.maskIRQ1, c.irq.maskTimer = s.IRQ.MaskIRQ2, s.IRQ.MaskIRQ1, s.IRQ.MaskTimer
	c.irq.pendingIRQ2, c.irq.pendingIRQ1, c.irq.pendingTimer = s.IRQ.PendingIRQ2, s.IRQ.PendingIRQ1, s.IRQ.PendingTimer
	c.irq.timerReload, c.irq.timerCounter = s.IRQ.TimerReload, s.IRQ.TimerCounter
	c.irq.timerAccum, c.irq.timerRunning = s.IRQ.TimerAccum, s.IRQ.TimerRunning
	c.pendingNMI = s.PendingNMI
}
