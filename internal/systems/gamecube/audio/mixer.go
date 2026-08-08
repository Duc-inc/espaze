// Package audio implements a simplified stand-in for the GameCube's
// audio hardware: a multi-channel mixer, not the real "DSP"
// coprocessor. The real DSP is its own custom-instruction-set
// processor (loosely comparable in obscurity to the SNES's SPC700 -
// see internal/systems/snes/spc700's own doc comment for why this
// project doesn't attempt a from-scratch interpreter for chips this
// under-documented) that runs real DSP microcode to drive its own
// ADPCM decoding and mixing. This project skips the DSP CPU entirely,
// but does decode real GameCube ADPCM data directly (see the sibling
// adpcm package and SetADPCMChannel below) as well as accepting
// already-decoded PCM (SetChannel) - the DSP microcode driving that
// decoding, and its own mixing hardware, aren't modeled.
package audio

import "github.com/Duc-inc/espaze/internal/systems/gamecube/adpcm"

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
	adpcm   *adpcmSource // nil unless SetADPCMChannel installed one
}

// adpcmSource decodes one 8-byte ADPCM frame's worth of samples at a
// time from compressed data, doling them out one per nextSample call
// (Step) - the actual "streaming" real hardware's DSP would do
// continuously as it plays.
type adpcmSource struct {
	decoder *adpcm.Decoder
	data    []byte
	pos     int
	buf     [14]int16
	bufLen  int
	bufPos  int
}

func (a *adpcmSource) nextSample() int16 {
	if a.bufPos >= a.bufLen {
		if a.pos+8 > len(a.data) {
			return 0 // stream exhausted
		}
		var frame [8]byte
		copy(frame[:], a.data[a.pos:a.pos+8])
		a.buf = a.decoder.DecodeFrame(frame)
		a.bufLen = len(a.buf)
		a.bufPos = 0
		a.pos += 8
	}
	s := a.buf[a.bufPos]
	a.bufPos++
	return s
}

// New returns a Mixer with every channel silent.
func New() *Mixer { return &Mixer{} }

// Reset silences every channel.
func (m *Mixer) Reset() { *m = Mixer{} }

// SetChannel feeds one channel's current sample and volume directly -
// for callers that already have decoded PCM, bypassing SetADPCMChannel's
// decoding entirely. Real software would refill this continuously as
// it streams audio.
func (m *Mixer) SetChannel(index int, sample int16, volume byte, enabled bool) {
	if index < 0 || index >= channelCount {
		return
	}
	m.channels[index] = channel{sample: sample, volume: volume, enabled: enabled}
}

// SetADPCMChannel points a channel at real GameCube ADPCM-compressed
// data plus the coefficient table it needs to decode it: Step
// advances through it 14 samples at a time (one 8-byte frame),
// feeding decoded PCM into this channel exactly like SetChannel would
// supply it manually.
func (m *Mixer) SetADPCMChannel(index int, coefs adpcm.Coefficients, data []byte, volume byte, enabled bool) {
	if index < 0 || index >= channelCount {
		return
	}
	m.channels[index] = channel{
		volume:  volume,
		enabled: enabled,
		adpcm:   &adpcmSource{decoder: adpcm.NewDecoder(coefs), data: data},
	}
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
// treated as "how many output samples to produce now". Any channel
// with an ADPCM source (SetADPCMChannel) decodes one more sample from
// it on every tick, exactly as if new PCM had arrived via SetChannel.
func (m *Mixer) Step(cycles int) {
	for i := 0; i < cycles; i++ {
		for c := range m.channels {
			if src := m.channels[c].adpcm; src != nil {
				m.channels[c].sample = src.nextSample()
			}
		}
		m.samples = append(m.samples, m.MixSample())
	}
}

// DrainSamples returns and clears every sample generated since the last call.
func (m *Mixer) DrainSamples() []int16 {
	out := m.samples
	m.samples = nil
	return out
}
