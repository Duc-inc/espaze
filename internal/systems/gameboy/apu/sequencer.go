package apu

// sequencerPeriod is how many T-cycles separate each 512Hz frame
// sequencer tick (4194304 / 512).
const sequencerPeriod = 8192

// frameSequencer derives the length counter (256Hz), sweep (128Hz) and
// volume envelope (64Hz) clocks from one 512Hz timer, the way real DMG
// hardware does, so every channel stays in lockstep with the others.
type frameSequencer struct {
	timer int
	step  int
}

// advance ticks the sequencer by cycles, calling back into apu for every
// 512Hz step it crosses (there can be more than one per call).
func (f *frameSequencer) advance(cycles int, a *APU) {
	f.timer += cycles
	for f.timer >= sequencerPeriod {
		f.timer -= sequencerPeriod
		a.tickSequencerStep(f.step)
		f.step = (f.step + 1) % 8
	}
}
