package apu

// SampleRate is the PCM rate every generated sample chunk is at.
const SampleRate = 44100

const cpuClockHz = 4194304.0
const cyclesPerSample = cpuClockHz / SampleRate

// APU is the DMG's 4-channel sound generator: two square waves, one
// user-defined waveform, and noise, mixed down to mono (the engine's
// audio pipeline is mono end to end - see internal/emulation/audio).
type APU struct {
	ch1 *squareChannel
	ch2 *squareChannel
	ch3 *waveChannel
	ch4 *noiseChannel

	seq frameSequencer

	masterLeft  byte
	masterRight byte
	panning     byte
	powerOn     bool

	sampleCycles float64
	samples      []int16
}

// New returns an APU that's powered off, matching the post-boot state
// CPU.Reset() will immediately write over via NR52.
func New() *APU {
	return &APU{
		ch1: newSquareChannel(true),
		ch2: newSquareChannel(false),
		ch3: newWaveChannel(),
		ch4: newNoiseChannel(),
	}
}

// Reset silences every channel without touching wave RAM (real hardware
// keeps it across a power cycle).
func (a *APU) Reset() {
	wave := a.ch3.ram
	*a = *New()
	a.ch3.ram = wave
}

// Step advances every channel and the frame sequencer by cycles, and
// appends however many audio samples fall due at 44.1kHz.
func (a *APU) Step(cycles int) {
	if !a.powerOn {
		return
	}

	// Advance in chunks that never cross a sample boundary, so a sample
	// always reflects the channels' state at the moment it was actually
	// due - not wherever they ended up after the whole cycles argument
	// was applied. Callers are expected to pass small per-instruction
	// counts anyway, but this keeps correctness independent of that.
	for cycles > 0 {
		chunk := cycles
		if untilNextSample := cyclesPerSample - a.sampleCycles; float64(chunk) > untilNextSample {
			chunk = int(untilNextSample)
			if chunk < 1 {
				chunk = 1
			}
		}

		a.ch1.step(chunk)
		a.ch2.step(chunk)
		a.ch3.step(chunk)
		a.ch4.step(chunk)
		a.seq.advance(chunk, a)
		cycles -= chunk

		a.sampleCycles += float64(chunk)
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

func (a *APU) tickSequencerStep(step int) {
	if step%2 == 0 {
		if a.ch1.length.tick() {
			a.ch1.enabled = false
		}
		if a.ch2.length.tick() {
			a.ch2.enabled = false
		}
		if a.ch3.length.tick() {
			a.ch3.enabled = false
		}
		if a.ch4.length.tick() {
			a.ch4.enabled = false
		}
	}
	if step == 2 || step == 6 {
		a.ch1.tickSweep()
	}
	if step == 7 {
		a.ch1.envelope.tick()
		a.ch2.envelope.tick()
		a.ch4.envelope.tick()
	}
}

// amplitudeScale converts one point of the averaged, centered channel
// output (-15..15) into int16 units. Chosen so the loudest possible mix
// (all channels at max volume, master volume maxed) stays inside int16's
// range with headroom to spare.
const amplitudeScale = 1800

// mixSample averages the DAC-enabled channels' raw 0-15 outputs (a
// channel with its DAC off is excluded entirely rather than contributing
// a false "0", which would bias the mix every time a channel is unused),
// centers that average around silence, and applies NR50's master volume
// (we're mono, so the left/right knobs are just averaged together).
func (a *APU) mixSample() int16 {
	sum, count := 0, 0
	for _, active := range []struct {
		on   bool
		level byte
	}{
		{a.ch1.active(), a.ch1.output()},
		{a.ch2.active(), a.ch2.output()},
		{a.ch3.active(), a.ch3.output()},
		{a.ch4.active(), a.ch4.output()},
	} {
		if active.on {
			sum += int(active.level)
			count++
		}
	}
	if count == 0 {
		return 0
	}

	average := sum * 2 / count // 0-30, kept as an even scale for centering
	centered := average - 15   // -15..15

	masterVolume := (int(a.masterLeft) + int(a.masterRight) + 1) / 2 // 0-7
	sample := centered * amplitudeScale * (masterVolume + 1) / 8
	switch {
	case sample > 32767:
		return 32767
	case sample < -32768:
		return -32768
	default:
		return int16(sample)
	}
}
