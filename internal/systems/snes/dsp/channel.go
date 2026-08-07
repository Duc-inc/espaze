package dsp

// channel is one of the DSP's 8 voices: a 32-sample signed-8-bit
// wavetable played back at a rate set by its 14-bit pitch register.
type channel struct {
	enabled bool
	volume  byte
	pitch   uint16

	wave         [32]int8
	waveWriteIdx byte
	playIdx      byte
	phaseAccum   int
}

func (c *channel) keyOn() {
	c.enabled = true
	c.playIdx = 0
	c.waveWriteIdx = 0
}

func (c *channel) keyOff() { c.enabled = false }

func (c *channel) writePitchLow(v byte)  { c.pitch = c.pitch&0x3F00 | uint16(v) }
func (c *channel) writePitchHigh(v byte) { c.pitch = c.pitch&0x00FF | uint16(v&0x3F)<<8 }

func (c *channel) writeWaveByte(v byte) {
	c.wave[c.waveWriteIdx&0x1F] = int8(v)
	c.waveWriteIdx++
}

const periodScale = 8

func (c *channel) tick() {
	if !c.enabled || c.pitch == 0 {
		return
	}
	c.phaseAccum++
	period := (0x4000 / int(c.pitch)) * periodScale
	if period < 1 {
		period = 1
	}
	if c.phaseAccum >= period {
		c.phaseAccum = 0
		c.playIdx = (c.playIdx + 1) % 32
	}
}

func (c *channel) sample() int16 {
	if !c.enabled {
		return 0
	}
	return int16(c.wave[c.playIdx]) * int16(c.volume)
}
