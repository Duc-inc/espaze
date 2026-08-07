package tia

// audioChannel is a deliberately simplified model of one of the TIA's
// two sound channels. Real hardware drives a 4-bit or 5-bit polynomial
// counter (a different tap/length per AUDC value, several of which
// produce very particular buzzy/noisy timbres) clocked by AUDF; this
// instead picks between a square wave (for the "tone-like" AUDC
// values) and a single shared LFSR noise generator (for the rest) -
// close enough to produce a recognizable Atari-era bleep/noise, not a
// bit-exact reproduction of any specific AUDC mode's waveform.
type audioChannel struct {
	audc   byte
	audf   byte
	volume byte

	divCounter int
	lfsr       uint16
	square     bool
}

func newAudioChannel() audioChannel { return audioChannel{lfsr: 1} }

func (c *audioChannel) writeAUDC(v byte) { c.audc = v & 0x0F }
func (c *audioChannel) writeAUDF(v byte) { c.audf = v & 0x1F }
func (c *audioChannel) writeAUDV(v byte) { c.volume = v & 0x0F }

func (c *audioChannel) isToneLike() bool {
	switch c.audc {
	case 0, 1, 12, 14, 15:
		return true
	default:
		return false
	}
}

// tick advances the channel by one CPU cycle, toggling its waveform
// generator every (AUDF+1) cycles - an approximation of AUDF's real
// role dividing the shared audio clock.
func (c *audioChannel) tick() {
	c.divCounter++
	if c.divCounter <= int(c.audf) {
		return
	}
	c.divCounter = 0

	if c.isToneLike() {
		c.square = !c.square
		return
	}

	bit := (c.lfsr ^ (c.lfsr >> 1)) & 1
	c.lfsr = (c.lfsr >> 1) | (bit << 14)
}

func (c *audioChannel) sample() int16 {
	if c.audc == 0 || c.volume == 0 {
		return 0
	}
	active := c.square
	if !c.isToneLike() {
		active = c.lfsr&1 != 0
	}
	if !active {
		return 0
	}
	return int16(c.volume) * 400
}
