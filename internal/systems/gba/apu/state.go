package apu

// Snapshot captures the APU's FIFOs and mix settings.
type Snapshot struct {
	FIFOA, FIFOB     []int8
	LastA, LastB     int8
	VolA, VolB       bool
	EnableA, EnableB bool
	DrainCycles      float64
	SampleCycles     float64
}

// Snapshot captures the APU's current state.
func (a *APU) Snapshot() Snapshot {
	fifoA := append([]int8(nil), a.fifoA...)
	fifoB := append([]int8(nil), a.fifoB...)
	return Snapshot{
		FIFOA: fifoA, FIFOB: fifoB, LastA: a.lastA, LastB: a.lastB,
		VolA: a.volA, VolB: a.volB, EnableA: a.enableA, EnableB: a.enableB,
		DrainCycles: a.drainCycles, SampleCycles: a.sampleCycles,
	}
}

// Restore reinstates a previously captured Snapshot.
func (a *APU) Restore(s Snapshot) {
	a.fifoA = append([]int8(nil), s.FIFOA...)
	a.fifoB = append([]int8(nil), s.FIFOB...)
	a.lastA, a.lastB = s.LastA, s.LastB
	a.volA, a.volB, a.enableA, a.enableB = s.VolA, s.VolB, s.EnableA, s.EnableB
	a.drainCycles, a.sampleCycles = s.DrainCycles, s.SampleCycles
}
