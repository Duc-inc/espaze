package apu

// waveChannel implements channel 3: playback of a user-defined 32-sample
// waveform (4 bits each) stored in wave RAM, at a programmable frequency.
type waveChannel struct {
	length lengthCounter

	dacEnabled bool
	volumeCode byte // 0=mute, 1=100%, 2=50%, 3=25%
	frequency  uint16
	timer      int
	position   byte
	ram        [16]byte
	enabled    bool
}

func newWaveChannel() *waveChannel {
	return &waveChannel{length: newLengthCounter(256)}
}

func (c *waveChannel) periodCycles() int {
	return (2048 - int(c.frequency)) * 2
}

func (c *waveChannel) step(cycles int) {
	if !c.enabled {
		return
	}
	c.timer -= cycles
	for c.timer <= 0 {
		c.timer += c.periodCycles()
		c.position = (c.position + 1) % 32
	}
}

func (c *waveChannel) currentSample() byte {
	b := c.ram[c.position/2]
	if c.position%2 == 0 {
		return b >> 4
	}
	return b & 0x0F
}

// active reports whether this channel should count toward the mix.
func (c *waveChannel) active() bool {
	return c.enabled && c.dacEnabled
}

func (c *waveChannel) output() byte {
	if !c.enabled || !c.dacEnabled {
		return 0
	}
	sample := c.currentSample()
	switch c.volumeCode {
	case 1:
		return sample
	case 2:
		return sample >> 1
	case 3:
		return sample >> 2
	default:
		return 0
	}
}

func (c *waveChannel) trigger() {
	c.enabled = c.dacEnabled
	if c.length.counter <= 0 {
		c.length.counter = c.length.max
	}
	c.timer = c.periodCycles()
	c.position = 0
}
