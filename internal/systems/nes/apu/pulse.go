package apu

// dutyTable holds the four selectable pulse waveforms (12.5%, 25%, 50%,
// 75% duty cycle), one bit per step out of 8.
var dutyTable = [4][8]byte{
	{0, 1, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 1, 0, 0, 0},
	{1, 0, 0, 1, 1, 1, 1, 1},
}

// pulseChannel implements APU channels 1 and 2: a square wave with
// selectable duty, a volume envelope, and a pitch-bending sweep unit.
type pulseChannel struct {
	env    envelope
	sweep  sweep
	length lengthCounter

	duty     byte
	dutyStep byte
	timer    uint16 // 11-bit period from $4002/3 or $4006/7
	timerCnt uint16
	enabled  bool
}

func newPulseChannel(onesComplement bool) *pulseChannel {
	return &pulseChannel{sweep: newSweep(onesComplement)}
}

// writeControl handles $4000/$4004.
func (c *pulseChannel) writeControl(v byte) {
	c.duty = v >> 6
	c.env.writeControl(v)
	c.length.halt = c.env.loop
}

func (c *pulseChannel) writeTimerLow(v byte) {
	c.timer = (c.timer &^ 0x00FF) | uint16(v)
}

// writeTimerHighAndLength handles $4003/$4007: the length-load index
// plus the timer's top 3 bits, and retriggers the envelope/duty phase.
func (c *pulseChannel) writeTimerHighAndLength(v byte) {
	c.timer = (c.timer &^ 0x0700) | (uint16(v&0x07) << 8)
	if c.enabled {
		c.length.load(v >> 3)
	}
	c.env.restart()
	c.dutyStep = 0
}

func (c *pulseChannel) setEnabled(on bool) {
	c.enabled = on
	if !on {
		c.length.value = 0
	}
}

func (c *pulseChannel) tickTimer() {
	if c.timerCnt == 0 {
		c.timerCnt = c.timer
		c.dutyStep = (c.dutyStep + 1) % 8
	} else {
		c.timerCnt--
	}
}

func (c *pulseChannel) tickSweep() { c.timer = c.sweep.tick(c.timer) }

func (c *pulseChannel) muted() bool {
	_, sweepMuted := c.sweep.target(c.timer)
	return sweepMuted || c.timer < 8
}

func (c *pulseChannel) active() bool { return c.enabled && c.length.active() }

func (c *pulseChannel) output() byte {
	if !c.active() || c.muted() || dutyTable[c.duty][c.dutyStep] == 0 {
		return 0
	}
	return c.env.volume()
}
