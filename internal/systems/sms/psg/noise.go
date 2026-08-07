package psg

// noiseChannel implements the PSG's 4th channel: a 16-bit
// linear-feedback shift register clocked either by one of 3 fixed
// periods or by tone channel 2's own flip-flop (shift rate 3) -
// "periodic" noise taps a single bit for a buzzy, tonal sound, "white"
// noise taps two for a hissier one.
type noiseChannel struct {
	shiftRate byte // 0-2 index into the fixed periods, 3 = clocked by tone2
	fbMode    bool // true = white noise, false = periodic
	lfsr      uint16
	counter   uint16
	atten     byte

	tone2           *toneChannel
	tone2PrevOutput bool
}

func newNoiseChannel(tone2 *toneChannel) *noiseChannel {
	return &noiseChannel{atten: 0x0F, lfsr: 0x8000, tone2: tone2}
}

// setControl handles a write to the noise control register - real
// hardware resets the LFSR whenever this register is written.
func (c *noiseChannel) setControl(v byte) {
	c.shiftRate = v & 0x03
	c.fbMode = v&0x04 != 0
	c.lfsr = 0x8000
}

func (c *noiseChannel) period() uint16 {
	switch c.shiftRate {
	case 0:
		return 512
	case 1:
		return 1024
	default:
		return 2048
	}
}

// tick must run after tone2.tick() in the same prescaled cycle, since
// shift-rate-3 mode detects tone2's transition within that same step.
func (c *noiseChannel) tick() {
	if c.shiftRate == 3 {
		justRoseHigh := c.tone2.output && !c.tone2PrevOutput
		c.tone2PrevOutput = c.tone2.output
		if justRoseHigh {
			c.clockLFSR()
		}
		return
	}
	if c.counter == 0 {
		c.counter = c.period()
		c.clockLFSR()
	} else {
		c.counter--
	}
}

func (c *noiseChannel) clockLFSR() {
	var feedback uint16
	if c.fbMode {
		feedback = (c.lfsr ^ (c.lfsr >> 3)) & 1
	} else {
		feedback = c.lfsr & 1
	}
	c.lfsr = c.lfsr>>1 | feedback<<15
}

func (c *noiseChannel) output() bool { return c.lfsr&1 != 0 }
