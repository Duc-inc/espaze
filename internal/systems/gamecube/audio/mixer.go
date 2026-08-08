// Package audio implements a simplified stand-in for the GameCube's
// audio hardware: a multi-channel PCM mixer, not the real "DSP"
// coprocessor. The real DSP is its own custom-instruction-set
// processor (loosely comparable in obscurity to the SNES's SPC700 -
// see internal/systems/snes/spc700's own doc comment for why this
// project doesn't attempt a from-scratch interpreter for chips this
// under-documented) that decodes ADPCM-compressed samples and mixes
// them per real hardware's own DSP microcode. This project skips the
// DSP CPU entirely and exposes its channels directly as signed
// 16-bit PCM sources instead - a bigger simplification than this
// project's other audio chips, appropriate for groundwork that isn't
// wired into a playable system anyway.
package audio

const SampleRate = 44100
const channelCount = 16 // GameCube hardware mixes up to 16 simultaneous voices

// Mixer holds every channel's current sample and volume.
type Mixer struct {
	channels [channelCount]channel
	samples  []int16
}

type channel struct {
	sample  int16
	volume  byte // 0-255
	enabled bool
}

// New returns a Mixer with every channel silent.
func New() *Mixer { return &Mixer{} }

// Reset silences every channel.
func (m *Mixer) Reset() { *m = Mixer{} }

// SetChannel feeds one channel's current sample and volume - real
// software would refill this continuously as it streams decoded
// audio; this project has no decoder, so callers provide already-PCM
// data.
func (m *Mixer) SetChannel(index int, sample int16, volume byte, enabled bool) {
	if index < 0 || index >= channelCount {
		return
	}
	m.channels[index] = channel{sample: sample, volume: volume, enabled: enabled}
}

// MixSample sums every enabled channel's current sample, scaled by
// its volume, into one output sample - called once per output sample
// period by whatever drives this mixer's timing.
func (m *Mixer) MixSample() int16 {
	var sum int32
	for i := range m.channels {
		ch := &m.channels[i]
		if !ch.enabled {
			continue
		}
		sum += int32(ch.sample) * int32(ch.volume) / 255
	}
	switch {
	case sum > 32767:
		return 32767
	case sum < -32768:
		return -32768
	default:
		return int16(sum)
	}
}

// Step generates cycles worth of samples at SampleRate - since this
// project doesn't tie the mixer to a real DSP clock, cycles is simply
// treated as "how many output samples to produce now".
func (m *Mixer) Step(cycles int) {
	for i := 0; i < cycles; i++ {
		m.samples = append(m.samples, m.MixSample())
	}
}

// DrainSamples returns and clears every sample generated since the last call.
func (m *Mixer) DrainSamples() []int16 {
	out := m.samples
	m.samples = nil
	return out
}
