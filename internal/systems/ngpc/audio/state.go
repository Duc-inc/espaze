package audio

// Snapshot captures the coprocessor's own RAM (the PSG is snapshotted
// separately by its owner).
type Snapshot struct {
	RAM [0x4000]byte
}

// Snapshot captures the Z80 bus's current state.
func (b *Bus) Snapshot() Snapshot { return Snapshot{RAM: b.ram} }

// Restore reinstates a previously captured Snapshot.
func (b *Bus) Restore(s Snapshot) { b.ram = s.RAM }
