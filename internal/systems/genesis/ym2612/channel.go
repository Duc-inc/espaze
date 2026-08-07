package ym2612

// channel is one of the YM2612's 6 FM voices: 4 operators combined
// according to one of 8 "algorithms" (how they're wired together -
// chained as modulator->carrier, or run in parallel as independent
// carriers, or something between). This implements 3 representative
// wiring shapes rather than all 8 exactly: full serial chain
// (algorithms 0-3), two parallel 2-operator chains (4-6), and full
// parallel/additive (7) - a deliberate simplification of the real
// per-algorithm operator graph, honest rather than silently wrong,
// that still produces recognizably different FM timbres.
type channel struct {
	ops [4]operator

	feedback  byte // op1's self-feedback amount, 0-7
	algorithm byte // 0-7

	fnum  uint16 // 11-bit frequency number
	block byte   // 3-bit octave/block
	keyOn [4]bool

	leftOn, rightOn bool

	feedbackHistory float64
}

func (c *channel) writeFeedbackAlgorithm(v byte) {
	c.feedback = (v >> 3) & 0x07
	c.algorithm = v & 0x07
}

func (c *channel) writeFNumLow(v byte) {
	c.fnum = c.fnum&0x0700 | uint16(v)
}

func (c *channel) writeFNumHighBlock(v byte) {
	c.fnum = c.fnum&0x00FF | uint16(v&0x07)<<8
	c.block = (v >> 3) & 0x07
}

func (c *channel) writePan(v byte) {
	c.leftOn = v&0x80 != 0
	c.rightOn = v&0x40 != 0
}

// baseFrequency converts F-NUM/block into Hz using the standard OPN2
// formula.
func (c *channel) baseFrequency() float64 {
	return float64(c.fnum) * (baseClock / 144.0) * pow2(int(c.block)-1)
}

const baseClock = 7670000.0 / 6 / 6 // approximate YM2612 internal clock division

func pow2(n int) float64 {
	if n >= 0 {
		return float64(uint32(1) << uint(n))
	}
	return 1.0 / float64(uint32(1)<<uint(-n))
}

func (c *channel) keyOnOp(opIdx byte) {
	c.ops[opIdx].keyOnEvent()
}
func (c *channel) keyOffOp(opIdx byte) {
	c.ops[opIdx].keyOffEvent()
}

// step produces one sample by running all 4 operators through this
// channel's algorithm shape.
func (c *channel) step() float64 {
	freq := c.baseFrequency()
	fbIn := c.feedbackHistory * feedbackScale(c.feedback)

	var out float64
	switch {
	case c.algorithm <= 3: // full serial chain: 1 -> 2 -> 3 -> 4
		o1 := c.ops[0].step(freq, fbIn)
		o2 := c.ops[1].step(freq, o1*modDepth)
		o3 := c.ops[2].step(freq, o2*modDepth)
		o4 := c.ops[3].step(freq, o3*modDepth)
		out = o4
		c.feedbackHistory = o1
	case c.algorithm <= 6: // two parallel chains: (1->2) + (3->4)
		o1 := c.ops[0].step(freq, fbIn)
		o2 := c.ops[1].step(freq, o1*modDepth)
		o3 := c.ops[2].step(freq, 0)
		o4 := c.ops[3].step(freq, o3*modDepth)
		out = (o2 + o4) / 2
		c.feedbackHistory = o1
	default: // 7: fully parallel/additive
		o1 := c.ops[0].step(freq, fbIn)
		o2 := c.ops[1].step(freq, 0)
		o3 := c.ops[2].step(freq, 0)
		o4 := c.ops[3].step(freq, 0)
		out = (o1 + o2 + o3 + o4) / 4
		c.feedbackHistory = o1
	}
	return out
}

const modDepth = 2.0 // radians of phase shift a full-amplitude modulator applies

func feedbackScale(fb byte) float64 {
	if fb == 0 {
		return 0
	}
	return float64(uint16(1)<<uint(fb)) / 256.0
}
