package apu

// dmcRateTable is the 16 selectable timer periods (in CPU cycles) for
// how often a new output bit gets clocked out.
var dmcRateTable = [16]uint16{
	428, 380, 340, 320, 286, 254, 226, 214, 190, 160, 142, 128, 106, 84, 72, 54,
}

// MemoryReader is the CPU-visible memory the DMC channel streams its
// delta-encoded samples from - real hardware even briefly stalls the
// CPU to do this, which this simplified version doesn't reproduce.
type MemoryReader interface {
	Read(addr uint16) byte
}

// dmcChannel implements the delta modulation channel: it plays back a
// stream of 1-bit deltas fetched directly from CPU memory, nudging its
// output level up or down two units per bit rather than being driven by
// a waveform table like the other channels.
type dmcChannel struct {
	mem MemoryReader

	irqEnable bool
	loop      bool
	rate      uint16
	rateCnt   uint16

	level byte

	sampleAddr   uint16
	sampleLength uint16
	currentAddr  uint16
	bytesLeft    uint16

	sampleBuffer byte
	bufferFull   bool
	shiftReg     byte
	bitsLeft     byte
	silence      bool

	irqFlag bool
}

func newDMCChannel(mem MemoryReader) *dmcChannel {
	return &dmcChannel{mem: mem, silence: true}
}

// writeControl handles $4010.
func (c *dmcChannel) writeControl(v byte) {
	c.irqEnable = v&0x80 != 0
	c.loop = v&0x40 != 0
	c.rate = dmcRateTable[v&0x0F]
	if !c.irqEnable {
		c.irqFlag = false
	}
}

func (c *dmcChannel) writeDirectLoad(v byte)   { c.level = v & 0x7F }
func (c *dmcChannel) writeSampleAddr(v byte)   { c.sampleAddr = 0xC000 + uint16(v)*64 }
func (c *dmcChannel) writeSampleLength(v byte) { c.sampleLength = uint16(v)*16 + 1 }

// setEnabled implements $4015's DMC bit: disabling stops playback
// immediately; enabling (re)starts it only if it isn't already playing.
func (c *dmcChannel) setEnabled(on bool) {
	if !on {
		c.bytesLeft = 0
		return
	}
	if c.bytesLeft == 0 {
		c.currentAddr = c.sampleAddr
		c.bytesLeft = c.sampleLength
	}
}

func (c *dmcChannel) active() bool { return c.bytesLeft > 0 }

func (c *dmcChannel) tick() {
	if c.rateCnt == 0 {
		c.rateCnt = c.rate
		c.clockOutputUnit()
	} else {
		c.rateCnt--
	}
}

func (c *dmcChannel) clockOutputUnit() {
	if !c.bufferFull && c.bytesLeft > 0 {
		c.sampleBuffer = c.mem.Read(c.currentAddr)
		c.bufferFull = true
		c.currentAddr++
		if c.currentAddr == 0 {
			c.currentAddr = 0x8000
		}
		c.bytesLeft--
		if c.bytesLeft == 0 {
			if c.loop {
				c.currentAddr = c.sampleAddr
				c.bytesLeft = c.sampleLength
			} else if c.irqEnable {
				c.irqFlag = true
			}
		}
	}

	if c.bitsLeft == 0 {
		c.bitsLeft = 8
		if c.bufferFull {
			c.shiftReg = c.sampleBuffer
			c.bufferFull = false
			c.silence = false
		} else {
			c.silence = true
		}
	}

	if !c.silence {
		if c.shiftReg&1 != 0 {
			if c.level <= 125 {
				c.level += 2
			}
		} else if c.level >= 2 {
			c.level -= 2
		}
	}
	c.shiftReg >>= 1
	c.bitsLeft--
}

func (c *dmcChannel) output() byte { return c.level }
