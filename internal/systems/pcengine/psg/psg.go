// Package psg implements the PC Engine's built-in sound generator: 6
// wavetable channels (each with its own 32-sample 4-bit user-defined
// waveform), the last two of which can switch to a noise generator
// instead. Real hardware also has an LFO frequency-modulation mode
// tying channel 1 to channel 2, which this project doesn't implement -
// no game this project targets is known to rely on it for its core
// audio, and it's a rarely-used effect even among commercial titles.
package psg

// SampleRate is the PCM rate every generated sample chunk is at.
const SampleRate = 44100

const cpuClockHz = 7159090.0
const cyclesPerSample = cpuClockHz / SampleRate

// PSG holds all 6 channels plus the shared main-volume registers.
type PSG struct {
	selected byte
	channels [6]channel

	mainVolL, mainVolR byte

	sampleCycles float64
	samples      []int16
}

// New returns a PSG with every channel silent.
func New() *PSG { return &PSG{} }

// Reset silences every channel.
func (p *PSG) Reset() { *p = PSG{} }

// SelectChannel implements the channel-select port ($00): every
// following register write targets this channel until changed.
func (p *PSG) SelectChannel(v byte) { p.selected = v & 0x07 }

// WriteMainVolume implements the shared left/right main volume ports.
func (p *PSG) WriteMainVolumeLeft(v byte)  { p.mainVolL = v & 0x0F }
func (p *PSG) WriteMainVolumeRight(v byte) { p.mainVolR = v & 0x0F }

func (p *PSG) selectedChannel() *channel {
	if p.selected >= 6 {
		return &p.channels[0]
	}
	return &p.channels[p.selected]
}

// WriteFreqLow/WriteFreqHigh/WriteControl/WritePan/WriteWaveData/
// WriteNoiseControl target the currently selected channel's own
// registers.
func (p *PSG) WriteFreqLow(v byte)  { p.selectedChannel().writeFreqLow(v) }
func (p *PSG) WriteFreqHigh(v byte) { p.selectedChannel().writeFreqHigh(v) }
func (p *PSG) WriteControl(v byte)  { p.selectedChannel().writeControl(v) }
func (p *PSG) WritePan(v byte)      { p.selectedChannel().pan = v }
func (p *PSG) WriteWaveData(v byte) { p.selectedChannel().writeWaveData(v) }

// WriteNoiseControl only has an effect on channels 4/5 (index 4-5),
// matching real hardware.
func (p *PSG) WriteNoiseControl(v byte) {
	if p.selected >= 4 {
		p.selectedChannel().writeNoiseControl(v)
	}
}

// Step advances every channel and generates samples, interleaved
// cycle-by-cycle so a single large Step call still produces a
// correctly varying waveform.
func (p *PSG) Step(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		for c := range p.channels {
			p.channels[c].tick()
		}
		p.sampleCycles++
		if p.sampleCycles >= cyclesPerSample {
			p.sampleCycles -= cyclesPerSample
			p.samples = append(p.samples, p.mixSample())
		}
	}
}

func (p *PSG) mixSample() int16 {
	var sum int32
	for c := range p.channels {
		sum += int32(p.channels[c].sample())
	}
	scaled := sum / 3
	switch {
	case scaled > 32767:
		return 32767
	case scaled < -32768:
		return -32768
	default:
		return int16(scaled)
	}
}

// DrainSamples returns and clears every sample generated since the last call.
func (p *PSG) DrainSamples() []int16 {
	out := p.samples
	p.samples = nil
	return out
}
