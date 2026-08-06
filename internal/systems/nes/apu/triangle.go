package apu

// triangleSequence is the 32-step waveform the triangle channel cycles
// through - a linear ramp down from 15 to 0, then back up to 15.
var triangleSequence = [32]byte{
	15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0,
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
}

// triangleChannel implements APU channel 3: a fixed-volume triangle
// wave gated by both a length counter and a linear counter (the latter
// gives finer, per-frame control some games use for smoother notes).
type triangleChannel struct {
	length lengthCounter

	linearPeriod byte
	linearValue  byte
	linearReload bool
	control      bool // also length.halt's source, and disables the reload flag's auto-clear

	timer    uint16 // 11-bit period
	timerCnt uint16
	step     byte
	enabled  bool
}

// writeLinear handles $4008.
func (c *triangleChannel) writeLinear(v byte) {
	c.control = v&0x80 != 0
	c.linearPeriod = v & 0x7F
	c.length.halt = c.control
}

func (c *triangleChannel) writeTimerLow(v byte) {
	c.timer = (c.timer &^ 0x00FF) | uint16(v)
}

// writeTimerHighAndLength handles $400B.
func (c *triangleChannel) writeTimerHighAndLength(v byte) {
	c.timer = (c.timer &^ 0x0700) | (uint16(v&0x07) << 8)
	if c.enabled {
		c.length.load(v >> 3)
	}
	c.linearReload = true
}

func (c *triangleChannel) setEnabled(on bool) {
	c.enabled = on
	if !on {
		c.length.value = 0
	}
}

func (c *triangleChannel) tickLinear() {
	if c.linearReload {
		c.linearValue = c.linearPeriod
	} else if c.linearValue > 0 {
		c.linearValue--
	}
	if !c.control {
		c.linearReload = false
	}
}

func (c *triangleChannel) active() bool { return c.enabled && c.length.active() }

// tickTimer advances the sequencer, but only while both the length and
// linear counters are non-zero - real hardware keeps the timer itself
// running regardless (it can produce inaudible ultrasonic output at
// very low periods), which this simplification doesn't reproduce.
func (c *triangleChannel) tickTimer() {
	if !c.active() || c.linearValue == 0 {
		return
	}
	if c.timerCnt == 0 {
		c.timerCnt = c.timer
		c.step = (c.step + 1) % 32
	} else {
		c.timerCnt--
	}
}

func (c *triangleChannel) output() byte {
	if !c.active() {
		return 0
	}
	return triangleSequence[c.step]
}
