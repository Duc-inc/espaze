package apu

// noisePeriodTable is the 16 selectable timer periods (in APU cycles),
// fixed by hardware rather than computed from a formula.
var noisePeriodTable = [16]uint16{
	4, 8, 16, 32, 64, 96, 128, 160, 202, 254, 380, 508, 762, 1016, 1524, 2034,
}

// noiseChannel implements APU channel 4: pseudo-random noise from a
// 15-bit linear-feedback shift register, with the same envelope/length
// gating as the pulse channels.
type noiseChannel struct {
	env    envelope
	length lengthCounter

	modeShort bool // true = 93-step "metallic" tap, false = 32767-step
	period    uint16
	timerCnt  uint16
	lfsr      uint16
	enabled   bool
}

func newNoiseChannel() *noiseChannel {
	return &noiseChannel{lfsr: 1}
}

// writeControl handles $400C (identical layout to the pulse channels').
func (c *noiseChannel) writeControl(v byte) {
	c.env.writeControl(v)
	c.length.halt = c.env.loop
}

// writeMode handles $400E.
func (c *noiseChannel) writeMode(v byte) {
	c.modeShort = v&0x80 != 0
	c.period = noisePeriodTable[v&0x0F]
}

// writeLength handles $400F.
func (c *noiseChannel) writeLength(v byte) {
	if c.enabled {
		c.length.load(v >> 3)
	}
	c.env.restart()
}

func (c *noiseChannel) setEnabled(on bool) {
	c.enabled = on
	if !on {
		c.length.value = 0
	}
}

func (c *noiseChannel) tickTimer() {
	if c.timerCnt == 0 {
		c.timerCnt = c.period
		tap := byte(1)
		if c.modeShort {
			tap = 6
		}
		feedback := (c.lfsr ^ (c.lfsr >> tap)) & 1
		c.lfsr >>= 1
		c.lfsr |= feedback << 14
	} else {
		c.timerCnt--
	}
}

func (c *noiseChannel) active() bool { return c.enabled && c.length.active() }

func (c *noiseChannel) output() byte {
	if !c.active() || c.lfsr&1 != 0 {
		return 0
	}
	return c.env.volume()
}
