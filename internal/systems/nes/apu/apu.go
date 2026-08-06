package apu

// SampleRate is the PCM rate every generated sample chunk is at.
const SampleRate = 44100

const cpuClockHz = 1789773.0 // NTSC 2A03 clock
const cyclesPerSample = cpuClockHz / SampleRate

// APU is the NES's 5-channel sound generator: two pulse waves, one
// triangle, noise, and delta-modulated sample playback, mixed down to
// mono through the same non-linear formula real hardware's analog mixer
// approximates (see mixSample).
type APU struct {
	pulse1   *pulseChannel
	pulse2   *pulseChannel
	triangle *triangleChannel
	noise    *noiseChannel
	dmc      *dmcChannel
	seq      frameSequencer

	cycleParity bool // pulse/noise/DMC timers tick every other CPU cycle; triangle ticks every cycle

	sampleCycles float64
	samples      []int16
}

// New returns an APU wired to mem for the DMC channel's sample fetches.
func New(mem MemoryReader) *APU {
	return &APU{
		pulse1:   newPulseChannel(true),
		pulse2:   newPulseChannel(false),
		triangle: &triangleChannel{},
		noise:    newNoiseChannel(),
		dmc:      newDMCChannel(mem),
	}
}

// Reset silences every channel, keeping the DMC's memory link.
func (a *APU) Reset() {
	mem := a.dmc.mem
	*a = *New(mem)
}

// SetMemory (re)wires the DMC channel's sample source - useful when the
// bus it needs to read from doesn't exist yet at APU construction time
// (the bus itself depends on the APU, for register I/O).
func (a *APU) SetMemory(mem MemoryReader) { a.dmc.mem = mem }

// Step advances every channel and the frame sequencer by cpuCycles CPU
// cycles, interleaved with sample generation so a large Step call still
// produces a correctly varying waveform instead of freezing on the
// state reached after the last cycle.
func (a *APU) Step(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		a.seq.advance(1, a)

		a.triangle.tickTimer()
		if a.cycleParity {
			a.pulse1.tickTimer()
			a.pulse2.tickTimer()
			a.noise.tickTimer()
			a.dmc.tick()
		}
		a.cycleParity = !a.cycleParity

		a.sampleCycles++
		if a.sampleCycles >= cyclesPerSample {
			a.sampleCycles -= cyclesPerSample
			a.samples = append(a.samples, a.mixSample())
		}
	}
}

// DrainSamples returns and clears every sample generated since the last call.
func (a *APU) DrainSamples() []int16 {
	out := a.samples
	a.samples = nil
	return out
}

// IRQPending reports whether either IRQ source (the frame sequencer or
// the DMC) currently wants the CPU's attention.
func (a *APU) IRQPending() bool { return a.seq.irqFlag || a.dmc.irqFlag }

func (a *APU) tickQuarterFrame() {
	a.pulse1.env.tick()
	a.pulse2.env.tick()
	a.noise.env.tick()
	a.triangle.tickLinear()
}

func (a *APU) tickHalfFrame() {
	a.pulse1.length.tick()
	a.pulse2.length.tick()
	a.triangle.length.tick()
	a.noise.length.tick()
	a.pulse1.tickSweep()
	a.pulse2.tickSweep()
}

// mixSample applies the NES's own non-linear mixing formula (see
// https://www.nesdev.org/wiki/APU_Mixer) rather than a simple average -
// the real hardware's analog mixer isn't linear, and games' volume
// balance is tuned assuming it isn't.
func (a *APU) mixSample() int16 {
	p1 := float64(a.pulse1.output())
	p2 := float64(a.pulse2.output())
	t := float64(a.triangle.output())
	n := float64(a.noise.output())
	d := float64(a.dmc.output())

	var pulseOut float64
	if p1+p2 > 0 {
		pulseOut = 95.88 / (8128/(p1+p2) + 100)
	}
	var tndOut float64
	if t > 0 || n > 0 || d > 0 {
		tndOut = 159.79 / (1/(t/8227+n/12241+d/22638) + 100)
	}

	sample := (pulseOut + tndOut) * 32767
	switch {
	case sample > 32767:
		return 32767
	case sample < -32768:
		return -32768
	default:
		return int16(sample)
	}
}
