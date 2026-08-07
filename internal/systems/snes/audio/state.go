package audio

// Snapshot captures the SPC700 bus's own state (ARAM plus the DSP
// address latch; the DSP and shared ports are snapshotted separately
// by their owners).
type Snapshot struct {
	ARAM    [0x10000]byte
	DSPAddr byte
}

// Snapshot captures the bus's current state.
func (b *Bus) Snapshot() Snapshot { return Snapshot{ARAM: b.aram, DSPAddr: b.dspAddr} }

// Restore reinstates a previously captured Snapshot.
func (b *Bus) Restore(s Snapshot) {
	b.aram = s.ARAM
	b.dspAddr = s.DSPAddr
}
