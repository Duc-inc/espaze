package riot

// Snapshot captures the RIOT's RAM and timer state.
type Snapshot struct {
	RAM          [128]byte
	TimerValue   byte
	TimerShift   uint
	TimerCounter int
	TimerStarted bool
}

// Snapshot captures the RIOT's current state.
func (r *RIOT) Snapshot() Snapshot {
	return Snapshot{
		RAM: r.ram, TimerValue: r.timerValue, TimerShift: r.timerShift,
		TimerCounter: r.timerCounter, TimerStarted: r.timerStarted,
	}
}

// Restore reinstates a previously captured Snapshot.
func (r *RIOT) Restore(s Snapshot) {
	r.ram = s.RAM
	r.timerValue = s.TimerValue
	r.timerShift = s.TimerShift
	r.timerCounter = s.TimerCounter
	r.timerStarted = s.TimerStarted
}
