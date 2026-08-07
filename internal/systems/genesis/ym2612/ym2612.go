package ym2612

// SampleRate is the PCM rate every generated sample chunk is at.
const SampleRate = 44100

// cyclesPerSample assumes the chip is stepped at the 68000's clock
// rate (the two share the same master clock division on real
// hardware); the caller is expected to pass 68000 cycles to Step.
const cpuClockHz = 7670000.0
const cyclesPerSample = cpuClockHz / SampleRate

// YM2612 is a from-scratch, deliberately simplified implementation of
// the Genesis's FM sound chip: 6 channels, 4 operators each, register
// layout matching the real chip so games' actual register writes
// produce a recognizable FM voice - see channel.go for exactly which
// simplifications that involves.
type YM2612 struct {
	channels [6]channel

	addr1, addr2 byte

	sampleCycles float64
	samples      []int16
}

// New returns a YM2612 with every channel silent.
func New() *YM2612 { return &YM2612{} }

// Reset silences every channel.
func (y *YM2612) Reset() { *y = YM2612{} }

// Step advances the chip by cpuCycles 68000 cycles, generating samples
// interleaved with that advance so a single large Step call still
// produces a correctly varying waveform.
func (y *YM2612) Step(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		y.sampleCycles++
		if y.sampleCycles >= cyclesPerSample {
			y.sampleCycles -= cyclesPerSample
			y.samples = append(y.samples, y.mixSample())
		}
	}
}

// DrainSamples returns and clears every sample generated since the last call.
func (y *YM2612) DrainSamples() []int16 {
	out := y.samples
	y.samples = nil
	return out
}

func (y *YM2612) mixSample() int16 {
	sum := 0.0
	for i := range y.channels {
		ch := &y.channels[i]
		if !ch.leftOn && !ch.rightOn {
			continue
		}
		sum += ch.step()
	}

	scaled := sum / 6 * 20000
	switch {
	case scaled > 32767:
		return 32767
	case scaled < -32768:
		return -32768
	default:
		return int16(scaled)
	}
}
