package apu

// Frame-sequencer step boundaries, in CPU cycles from the start of the
// sequence (NTSC timing). Both modes share their first three points and
// only diverge from the 4th step onward.
const (
	seqStep1 = 7457
	seqStep2 = 14913
	seqStep3 = 22371
	seqStep4 = 29829
	seqStep5 = 37281
)

// frameSequencer drives the quarter-frame (envelopes, triangle linear
// counter) and half-frame (length counters, sweep units) clocks every
// channel but DMC depends on, in either 4-step mode (with an optional
// IRQ on the last step) or 5-step mode (no IRQ, one extra step).
type frameSequencer struct {
	cycle      int
	fiveStep   bool
	irqInhibit bool
	irqFlag    bool
}

// write handles $4017. Selecting 5-step mode clocks a quarter+half
// frame immediately, matching real hardware's "write resets and clocks".
func (f *frameSequencer) write(v byte, a *APU) {
	f.fiveStep = v&0x80 != 0
	f.irqInhibit = v&0x40 != 0
	if f.irqInhibit {
		f.irqFlag = false
	}
	f.cycle = 0
	if f.fiveStep {
		a.tickQuarterFrame()
		a.tickHalfFrame()
	}
}

// advance runs the sequencer forward by cpuCycles CPU cycles, clocking
// quarter/half-frame events on a whenever a step boundary is crossed.
func (f *frameSequencer) advance(cpuCycles int, a *APU) {
	for i := 0; i < cpuCycles; i++ {
		f.cycle++
		if f.fiveStep {
			f.tickFiveStep(a)
		} else {
			f.tickFourStep(a)
		}
	}
}

func (f *frameSequencer) tickFourStep(a *APU) {
	switch f.cycle {
	case seqStep1, seqStep3:
		a.tickQuarterFrame()
	case seqStep2:
		a.tickQuarterFrame()
		a.tickHalfFrame()
	case seqStep4:
		a.tickQuarterFrame()
		a.tickHalfFrame()
		if !f.irqInhibit {
			f.irqFlag = true
		}
		f.cycle = 0
	}
}

func (f *frameSequencer) tickFiveStep(a *APU) {
	switch f.cycle {
	case seqStep1, seqStep3:
		a.tickQuarterFrame()
	case seqStep2:
		a.tickQuarterFrame()
		a.tickHalfFrame()
	case seqStep5:
		a.tickQuarterFrame()
		a.tickHalfFrame()
		f.cycle = 0
	}
}
