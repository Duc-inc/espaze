package audio

// Snapshot captures the coprocessor's own RAM and control-line state
// (the YM2612 and PSG are snapshotted separately by their owners).
type Snapshot struct {
	RAM    [0x2000]byte
	Bank   uint16
	Halted bool
}

// Snapshot captures the Z80 bus's current state.
func (b *Bus) Snapshot() Snapshot {
	return Snapshot{RAM: b.ram, Bank: b.bank, Halted: b.halted}
}

// Restore reinstates a previously captured Snapshot.
func (b *Bus) Restore(s Snapshot) {
	b.ram = s.RAM
	b.bank = s.Bank
	b.halted = s.Halted
}
