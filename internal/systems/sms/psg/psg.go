package psg

// SampleRate is the PCM rate every generated sample chunk is at.
const SampleRate = 44100

const masterClockHz = 3579545.0 // NTSC SMS master clock, shared by the Z80 and PSG
const cyclesPerSample = masterClockHz / SampleRate

// PSG is a from-scratch implementation of the SN76489: 3 square-wave
// tone channels and 1 noise channel, each with its own 4-bit
// attenuation, mixed down to mono.
type PSG struct {
	tone0, tone1, tone2 *toneChannel
	noise               *noiseChannel

	latchedChannel byte
	latchedType    byte // 0 = frequency/control, 1 = attenuation

	prescaler    int
	sampleCycles float64
	samples      []int16
}

// New returns a PSG with every channel silent (max attenuation).
func New() *PSG {
	tone2 := newToneChannel()
	return &PSG{
		tone0: newToneChannel(),
		tone1: newToneChannel(),
		tone2: tone2,
		noise: newNoiseChannel(tone2),
	}
}

// Reset silences every channel.
func (p *PSG) Reset() { *p = *New() }

// Write implements the single PSG I/O port's two-byte-per-register
// protocol: a "latch" byte (bit7 set) selects a channel and register
// type and supplies its low 4 bits; for tone frequencies specifically, a
// following "data" byte (bit7 clear) supplies the high 6 bits.
func (p *PSG) Write(v byte) {
	if v&0x80 != 0 {
		p.latchedChannel = (v >> 5) & 0x03
		p.latchedType = (v >> 4) & 0x01
		p.applyData(v&0x0F, true)
	} else {
		p.applyData(v&0x3F, false)
	}
}

func (p *PSG) applyData(data byte, isLatch bool) {
	if p.latchedType == 1 {
		p.setAttenuation(data & 0x0F)
		return
	}
	switch p.latchedChannel {
	case 0:
		if isLatch {
			p.tone0.setFreqLow(data)
		} else {
			p.tone0.setFreqHigh(data)
		}
	case 1:
		if isLatch {
			p.tone1.setFreqLow(data)
		} else {
			p.tone1.setFreqHigh(data)
		}
	case 2:
		if isLatch {
			p.tone2.setFreqLow(data)
		} else {
			p.tone2.setFreqHigh(data)
		}
	default:
		p.noise.setControl(data)
	}
}

func (p *PSG) setAttenuation(v byte) {
	switch p.latchedChannel {
	case 0:
		p.tone0.atten = v
	case 1:
		p.tone1.atten = v
	case 2:
		p.tone2.atten = v
	default:
		p.noise.atten = v
	}
}

// Step advances every channel and generates samples, interleaved
// cycle-by-cycle rather than batched, so a single large Step call still
// produces a correctly varying waveform.
func (p *PSG) Step(cycles int) {
	for i := 0; i < cycles; i++ {
		p.prescaler++
		if p.prescaler >= 16 { // all 4 channels share one master-clock/16 prescaler
			p.prescaler = 0
			p.tone0.tick()
			p.tone1.tick()
			p.tone2.tick() // must run before noise.tick(): shift-rate-3 mode detects tone2's edge this same step
			p.noise.tick()
		}

		p.sampleCycles++
		if p.sampleCycles >= cyclesPerSample {
			p.sampleCycles -= cyclesPerSample
			p.samples = append(p.samples, p.mixSample())
		}
	}
}

// DrainSamples returns and clears every sample generated since the last call.
func (p *PSG) DrainSamples() []int16 {
	out := p.samples
	p.samples = nil
	return out
}

func (p *PSG) mixSample() int16 {
	sum := channelSample(p.tone0.output, p.tone0.atten) +
		channelSample(p.tone1.output, p.tone1.atten) +
		channelSample(p.tone2.output, p.tone2.atten) +
		channelSample(p.noise.output(), p.noise.atten)

	switch {
	case sum > 32767:
		return 32767
	case sum < -32768:
		return -32768
	default:
		return int16(sum)
	}
}

// channelSample treats a channel's square/noise output as a proper
// bipolar wave (+volume/-volume) rather than an on/off pulse, so a
// silent mix of all channels averages to zero instead of a DC offset.
func channelSample(high bool, atten byte) int32 {
	v := int32(attenToVolume(atten))
	if high {
		return v
	}
	return -v
}
