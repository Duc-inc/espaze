package ym2612

// operatorSnapshot captures one operator's registers plus its running
// phase/envelope, since gob silently drops unexported fields on any
// struct embedded directly rather than flattened like this.
type operatorSnapshot struct {
	Mul, Det, TL, AR, D1R, D2R, SL, RR byte
	Phase                              float64
	EnvPhase                           int
	EnvLevel                           float64
	KeyOn                              bool
}

func (op *operator) snapshot() operatorSnapshot {
	return operatorSnapshot{
		Mul: op.mul, Det: op.det, TL: op.tl, AR: op.ar,
		D1R: op.d1r, D2R: op.d2r, SL: op.sl, RR: op.rr,
		Phase: op.phase, EnvPhase: op.envPhase, EnvLevel: op.envLevel, KeyOn: op.keyOn,
	}
}

func (op *operator) restore(s operatorSnapshot) {
	op.mul, op.det, op.tl, op.ar = s.Mul, s.Det, s.TL, s.AR
	op.d1r, op.d2r, op.sl, op.rr = s.D1R, s.D2R, s.SL, s.RR
	op.phase, op.envPhase, op.envLevel, op.keyOn = s.Phase, s.EnvPhase, s.EnvLevel, s.KeyOn
}

type channelSnapshot struct {
	Ops             [4]operatorSnapshot
	Feedback        byte
	Algorithm       byte
	FNum            uint16
	Block           byte
	LeftOn, RightOn bool
	FeedbackHistory float64
}

func (c *channel) snapshot() channelSnapshot {
	var ops [4]operatorSnapshot
	for i := range c.ops {
		ops[i] = c.ops[i].snapshot()
	}
	return channelSnapshot{
		Ops: ops, Feedback: c.feedback, Algorithm: c.algorithm,
		FNum: c.fnum, Block: c.block,
		LeftOn: c.leftOn, RightOn: c.rightOn, FeedbackHistory: c.feedbackHistory,
	}
}

func (c *channel) restore(s channelSnapshot) {
	for i := range c.ops {
		c.ops[i].restore(s.Ops[i])
	}
	c.feedback, c.algorithm = s.Feedback, s.Algorithm
	c.fnum, c.block = s.FNum, s.Block
	c.leftOn, c.rightOn, c.feedbackHistory = s.LeftOn, s.RightOn, s.FeedbackHistory
}

// Snapshot captures the whole chip's state (not the pending sample
// buffer - like every other audio core here, that's transient).
type Snapshot struct {
	Channels     [6]channelSnapshot
	Addr1, Addr2 byte
	SampleCycles float64
}

// Snapshot captures the YM2612's current state.
func (y *YM2612) Snapshot() Snapshot {
	var chans [6]channelSnapshot
	for i := range y.channels {
		chans[i] = y.channels[i].snapshot()
	}
	return Snapshot{Channels: chans, Addr1: y.addr1, Addr2: y.addr2, SampleCycles: y.sampleCycles}
}

// Restore reinstates a previously captured Snapshot.
func (y *YM2612) Restore(s Snapshot) {
	for i := range y.channels {
		y.channels[i].restore(s.Channels[i])
	}
	y.addr1, y.addr2 = s.Addr1, s.Addr2
	y.sampleCycles = s.SampleCycles
}
