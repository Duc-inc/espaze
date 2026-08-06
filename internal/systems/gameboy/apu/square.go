package apu

// dutyTable holds the four selectable waveforms (12.5%, 25%, 50%, 75%
// duty cycle), one bit lit per step out of 8.
var dutyTable = [4][8]byte{
	{0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 1, 1, 1},
	{0, 1, 1, 1, 1, 1, 1, 0},
}

// squareChannel implements channels 1 and 2: a square wave with a
// selectable duty cycle, a volume envelope, and (channel 1 only) a
// frequency sweep.
type squareChannel struct {
	hasSweep bool
	sweep    frequencySweep
	length   lengthCounter
	envelope volumeEnvelope

	duty      byte
	dutyStep  byte
	frequency uint16
	timer     int
	enabled   bool
}

func newSquareChannel(hasSweep bool) *squareChannel {
	return &squareChannel{hasSweep: hasSweep, length: newLengthCounter(64)}
}

func (c *squareChannel) periodCycles() int {
	return (2048 - int(c.frequency)) * 4
}

func (c *squareChannel) step(cycles int) {
	if !c.enabled {
		return
	}
	c.timer -= cycles
	for c.timer <= 0 {
		c.timer += c.periodCycles()
		c.dutyStep = (c.dutyStep + 1) % 8
	}
}

// active reports whether this channel should count toward the mix at
// all - a triggered channel whose DAC is off contributes nothing, not a
// silent "0" (which would bias the average toward negative).
func (c *squareChannel) active() bool {
	return c.enabled && c.envelope.dacEnabled()
}

func (c *squareChannel) output() byte {
	if !c.enabled || !c.envelope.dacEnabled() {
		return 0
	}
	if dutyTable[c.duty][c.dutyStep] == 0 {
		return 0
	}
	return c.envelope.volume
}

func (c *squareChannel) trigger() {
	c.enabled = true
	if c.length.counter <= 0 {
		c.length.counter = c.length.max
	}
	c.timer = c.periodCycles()
	c.envelope.trigger()

	if c.hasSweep && c.sweep.trigger(c.frequency) {
		c.enabled = false
	}
	if !c.envelope.dacEnabled() {
		c.enabled = false
	}
}

func (c *squareChannel) tickSweep() {
	if !c.hasSweep {
		return
	}
	freq, apply, disable := c.sweep.tick()
	if disable {
		c.enabled = false
		return
	}
	if apply {
		c.frequency = freq
	}
}
