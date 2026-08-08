package memory

// Snapshot captures MEM1's contents.
type Snapshot struct {
	MEM1 [mem1Size]byte
}

// Snapshot captures the bus's current state.
func (b *Bus) Snapshot() Snapshot { return Snapshot{MEM1: b.mem1} }

// Restore reinstates a previously captured Snapshot.
func (b *Bus) Restore(s Snapshot) { b.mem1 = s.MEM1 }
