package apu

// noiseDivisors are the 8 selectable base periods (in T-cycles) for the
// noise channel's pseudo-random shift register, before the shift-clock
// exponent is applied.
var noiseDivisors = [8]int{8, 16, 32, 48, 64, 80, 96, 112}

// noiseChannel implements channel 4: pseudo-random noise from a
// linear-feedback shift register, with the same envelope every other
// channel but no game console(tm) DAC-shaped waveform - just static.
type noiseChannel struct {
	length   lengthCounter
	envelope volumeEnvelope

	shiftClockFreq byte
	widthMode      bool // true = 7-bit LFSR (metallic), false = 15-bit
	divisorCode    byte

	lfsr    uint16
	timer   int
	enabled bool
}

func newNoiseChannel() *noiseChannel {
	return &noiseChannel{length: newLengthCounter(64)}
}

func (c *noiseChannel) periodCycles() int {
	return noiseDivisors[c.divisorCode] << c.shiftClockFreq
}

func (c *noiseChannel) step(cycles int) {
	if !c.enabled {
		return
	}
	c.timer -= cycles
	for c.timer <= 0 {
		c.timer += c.periodCycles()
		bit := (c.lfsr ^ (c.lfsr >> 1)) & 1
		c.lfsr >>= 1
		c.lfsr |= bit << 14
		if c.widthMode {
			c.lfsr &^= 1 << 6
			c.lfsr |= bit << 6
		}
	}
}

// active reports whether this channel should count toward the mix.
func (c *noiseChannel) active() bool {
	return c.enabled && c.envelope.dacEnabled()
}

func (c *noiseChannel) output() byte {
	if !c.enabled || !c.envelope.dacEnabled() {
		return 0
	}
	if c.lfsr&1 != 0 {
		return 0
	}
	return c.envelope.volume
}

func (c *noiseChannel) trigger() {
	c.enabled = true
	if c.length.counter <= 0 {
		c.length.counter = c.length.max
	}
	c.lfsr = 0x7FFF
	c.envelope.trigger()
	if !c.envelope.dacEnabled() {
		c.enabled = false
	}
}
