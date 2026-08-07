// Package dsp implements a simplified slice of the SNES's S-DSP: 8
// channels, each with a volume, pitch, and a 32-sample looping
// wavetable the SPC700 driver writes directly (rather than real
// hardware's BRR - Bit Rate Reduction - compressed sample playback
// from ARAM, which decodes actual Nintendo-format compressed audio
// data; that decoder isn't implemented here). Echo, the noise
// generator, and per-channel ADSR envelopes aren't implemented
// either - every channel plays at a flat volume once keyed on.
package dsp

const SampleRate = 44100

const spcClockHz = 1024000.0 // the SPC700's own clock, close to real hardware
const cyclesPerSample = spcClockHz / SampleRate

// DSP holds all 8 channels.
type DSP struct {
	channels [8]channel

	sampleCycles float64
	samples      []int16
}

// New returns a DSP with every channel silent.
func New() *DSP { return &DSP{} }

// Reset silences every channel.
func (d *DSP) Reset() { *d = DSP{} }

// KeyOn/KeyOff start/stop a channel (0-7) playing its wavetable from
// the beginning.
func (d *DSP) KeyOn(ch int)  { d.channels[ch&0x07].keyOn() }
func (d *DSP) KeyOff(ch int) { d.channels[ch&0x07].keyOff() }

func (d *DSP) WriteVolume(ch int, v byte)    { d.channels[ch&0x07].volume = v }
func (d *DSP) WritePitchLow(ch int, v byte)  { d.channels[ch&0x07].writePitchLow(v) }
func (d *DSP) WritePitchHigh(ch int, v byte) { d.channels[ch&0x07].writePitchHigh(v) }
func (d *DSP) WriteWaveByte(ch int, v byte)  { d.channels[ch&0x07].writeWaveByte(v) }

// Step advances every channel and generates samples, interleaved
// cycle-by-cycle so a single large Step call still produces a
// correctly varying waveform.
func (d *DSP) Step(spcCycles int) {
	for i := 0; i < spcCycles; i++ {
		for c := range d.channels {
			d.channels[c].tick()
		}
		d.sampleCycles++
		if d.sampleCycles >= cyclesPerSample {
			d.sampleCycles -= cyclesPerSample
			d.samples = append(d.samples, d.mixSample())
		}
	}
}

func (d *DSP) mixSample() int16 {
	var sum int32
	for c := range d.channels {
		sum += int32(d.channels[c].sample())
	}
	scaled := sum / 4
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
func (d *DSP) DrainSamples() []int16 {
	out := d.samples
	d.samples = nil
	return out
}
