package psg

// toneChannel implements one of the PSG's 3 square-wave generators: a
// 10-bit frequency divider toggling a flip-flop, plus a 4-bit
// attenuation register (0 = loudest, 15 = silent - inverted from how
// "volume" usually reads, a real hardware quirk).
type toneChannel struct {
	freq    uint16
	counter uint16
	output  bool
	atten   byte
}

func newToneChannel() *toneChannel {
	return &toneChannel{atten: 0x0F}
}

func (c *toneChannel) setFreqLow(v byte)  { c.freq = (c.freq &^ 0x000F) | uint16(v&0x0F) }
func (c *toneChannel) setFreqHigh(v byte) { c.freq = (c.freq &^ 0x03F0) | (uint16(v&0x3F) << 4) }

// tick advances by one already-prescaled (master clock / 16) cycle.
func (c *toneChannel) tick() {
	if c.counter == 0 {
		c.counter = c.freq
		c.output = !c.output
	} else {
		c.counter--
	}
}
