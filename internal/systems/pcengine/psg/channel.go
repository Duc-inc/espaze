package psg

// channel is one of the PSG's 6 voices: a 32-sample, 4-bit
// user-defined waveform played back at a rate set by its 12-bit
// frequency register, or (channels 4-5 only) a simple LFSR noise
// generator instead. The exact real hardware constant relating the
// frequency register to playback rate isn't independently confirmed
// here - the scale factor below is this project's own tuning, chosen
// to produce audibly correct relative pitches rather than a verified
// absolute frequency.
type channel struct {
	freq    uint16
	enabled bool
	ddaMode bool
	volume  byte
	pan     byte

	wave           [32]byte
	waveWriteIndex byte
	playIndex      byte
	phaseAccum     int

	noiseMode    bool
	noiseFreq    byte
	noiseLFSR    uint32
	noiseCounter int
	noiseOutput  bool
}

func (c *channel) writeFreqLow(v byte)  { c.freq = c.freq&0xF00 | uint16(v) }
func (c *channel) writeFreqHigh(v byte) { c.freq = c.freq&0x0FF | uint16(v&0x0F)<<8 }

func (c *channel) writeControl(v byte) {
	wasDDA := c.ddaMode
	c.enabled = v&0x80 != 0
	c.ddaMode = v&0x40 != 0
	c.volume = v & 0x1F
	if c.ddaMode && !wasDDA {
		c.waveWriteIndex = 0
	}
}

func (c *channel) writeWaveData(v byte) {
	c.wave[c.waveWriteIndex&0x1F] = v & 0x0F
	c.waveWriteIndex++
}

func (c *channel) writeNoiseControl(v byte) {
	c.noiseMode = v&0x80 != 0
	c.noiseFreq = v & 0x1F
	if c.noiseLFSR == 0 {
		c.noiseLFSR = 1
	}
}

const periodScale = 16

func (c *channel) tick() {
	if !c.enabled {
		return
	}
	if c.noiseMode {
		c.noiseCounter++
		period := (int(c.noiseFreq) + 1) * periodScale
		if c.noiseCounter >= period {
			c.noiseCounter = 0
			bit := (c.noiseLFSR ^ (c.noiseLFSR >> 1)) & 1
			c.noiseLFSR = c.noiseLFSR>>1 | bit<<14
			c.noiseOutput = c.noiseLFSR&1 != 0
		}
		return
	}
	if c.ddaMode {
		return
	}
	c.phaseAccum++
	period := (int(c.freq) + 1) * periodScale
	if c.phaseAccum >= period {
		c.phaseAccum = 0
		c.playIndex = (c.playIndex + 1) % 32
	}
}

func (c *channel) sample() int16 {
	if !c.enabled {
		return 0
	}
	var raw byte
	if c.noiseMode {
		if c.noiseOutput {
			raw = 15
		}
	} else {
		raw = c.wave[c.playIndex]
	}
	centered := int32(raw) - 8
	return int16(centered * int32(c.volume) * 30)
}
