package apu

// The snapshot types below flatten each component's unexported fields
// into exported ones by hand, the same reasoning as the PPU's Snapshot:
// gob silently drops unexported struct fields rather than erroring, so
// embedding the internal types directly would quietly lose state.

type envelopeSnapshot struct {
	Period         byte
	ConstantVolume bool
	Loop           bool
	Start          bool
	Divider        byte
	DecayLevel     byte
}

func (e *envelope) snapshot() envelopeSnapshot {
	return envelopeSnapshot{e.period, e.constantVolume, e.loop, e.start, e.divider, e.decayLevel}
}

func (e *envelope) restore(s envelopeSnapshot) {
	e.period, e.constantVolume, e.loop = s.Period, s.ConstantVolume, s.Loop
	e.start, e.divider, e.decayLevel = s.Start, s.Divider, s.DecayLevel
}

type sweepSnapshot struct {
	Enabled bool
	Period  byte
	Negate  bool
	Shift   byte
	Reload  bool
	Divider byte
}

func (s *sweep) snapshot() sweepSnapshot {
	return sweepSnapshot{s.enabled, s.period, s.negate, s.shift, s.reload, s.divider}
}

func (s *sweep) restore(snap sweepSnapshot) {
	s.enabled, s.period, s.negate = snap.Enabled, snap.Period, snap.Negate
	s.shift, s.reload, s.divider = snap.Shift, snap.Reload, snap.Divider
}

type lengthSnapshot struct {
	Value byte
	Halt  bool
}

func (l *lengthCounter) snapshot() lengthSnapshot { return lengthSnapshot{l.value, l.halt} }
func (l *lengthCounter) restore(s lengthSnapshot) { l.value, l.halt = s.Value, s.Halt }

// PulseSnapshot captures one pulse channel's full state.
type PulseSnapshot struct {
	Env      envelopeSnapshot
	Sweep    sweepSnapshot
	Length   lengthSnapshot
	Duty     byte
	DutyStep byte
	Timer    uint16
	TimerCnt uint16
	Enabled  bool
}

func (c *pulseChannel) snapshot() PulseSnapshot {
	return PulseSnapshot{
		Env: c.env.snapshot(), Sweep: c.sweep.snapshot(), Length: c.length.snapshot(),
		Duty: c.duty, DutyStep: c.dutyStep, Timer: c.timer, TimerCnt: c.timerCnt, Enabled: c.enabled,
	}
}

func (c *pulseChannel) restore(s PulseSnapshot) {
	c.env.restore(s.Env)
	c.sweep.restore(s.Sweep)
	c.length.restore(s.Length)
	c.duty, c.dutyStep, c.timer, c.timerCnt, c.enabled = s.Duty, s.DutyStep, s.Timer, s.TimerCnt, s.Enabled
}

// TriangleSnapshot captures the triangle channel's full state.
type TriangleSnapshot struct {
	Length       lengthSnapshot
	LinearPeriod byte
	LinearValue  byte
	LinearReload bool
	Control      bool
	Timer        uint16
	TimerCnt     uint16
	Step         byte
	Enabled      bool
}

func (c *triangleChannel) snapshot() TriangleSnapshot {
	return TriangleSnapshot{
		Length: c.length.snapshot(), LinearPeriod: c.linearPeriod, LinearValue: c.linearValue,
		LinearReload: c.linearReload, Control: c.control,
		Timer: c.timer, TimerCnt: c.timerCnt, Step: c.step, Enabled: c.enabled,
	}
}

func (c *triangleChannel) restore(s TriangleSnapshot) {
	c.length.restore(s.Length)
	c.linearPeriod, c.linearValue, c.linearReload, c.control = s.LinearPeriod, s.LinearValue, s.LinearReload, s.Control
	c.timer, c.timerCnt, c.step, c.enabled = s.Timer, s.TimerCnt, s.Step, s.Enabled
}

// NoiseSnapshot captures the noise channel's full state.
type NoiseSnapshot struct {
	Env       envelopeSnapshot
	Length    lengthSnapshot
	ModeShort bool
	Period    uint16
	TimerCnt  uint16
	LFSR      uint16
	Enabled   bool
}

func (c *noiseChannel) snapshot() NoiseSnapshot {
	return NoiseSnapshot{
		Env: c.env.snapshot(), Length: c.length.snapshot(), ModeShort: c.modeShort,
		Period: c.period, TimerCnt: c.timerCnt, LFSR: c.lfsr, Enabled: c.enabled,
	}
}

func (c *noiseChannel) restore(s NoiseSnapshot) {
	c.env.restore(s.Env)
	c.length.restore(s.Length)
	c.modeShort, c.period, c.timerCnt, c.lfsr, c.enabled = s.ModeShort, s.Period, s.TimerCnt, s.LFSR, s.Enabled
}

// DMCSnapshot captures the delta-modulation channel's full state.
type DMCSnapshot struct {
	IrqEnable    bool
	Loop         bool
	Rate         uint16
	RateCnt      uint16
	Level        byte
	SampleAddr   uint16
	SampleLength uint16
	CurrentAddr  uint16
	BytesLeft    uint16
	SampleBuffer byte
	BufferFull   bool
	ShiftReg     byte
	BitsLeft     byte
	Silence      bool
	IrqFlag      bool
}

func (c *dmcChannel) snapshot() DMCSnapshot {
	return DMCSnapshot{
		IrqEnable: c.irqEnable, Loop: c.loop, Rate: c.rate, RateCnt: c.rateCnt, Level: c.level,
		SampleAddr: c.sampleAddr, SampleLength: c.sampleLength, CurrentAddr: c.currentAddr, BytesLeft: c.bytesLeft,
		SampleBuffer: c.sampleBuffer, BufferFull: c.bufferFull, ShiftReg: c.shiftReg,
		BitsLeft: c.bitsLeft, Silence: c.silence, IrqFlag: c.irqFlag,
	}
}

func (c *dmcChannel) restore(s DMCSnapshot) {
	c.irqEnable, c.loop, c.rate, c.rateCnt, c.level = s.IrqEnable, s.Loop, s.Rate, s.RateCnt, s.Level
	c.sampleAddr, c.sampleLength, c.currentAddr, c.bytesLeft = s.SampleAddr, s.SampleLength, s.CurrentAddr, s.BytesLeft
	c.sampleBuffer, c.bufferFull, c.shiftReg = s.SampleBuffer, s.BufferFull, s.ShiftReg
	c.bitsLeft, c.silence, c.irqFlag = s.BitsLeft, s.Silence, s.IrqFlag
}

// Snapshot captures the whole APU's state.
type Snapshot struct {
	Pulse1   PulseSnapshot
	Pulse2   PulseSnapshot
	Triangle TriangleSnapshot
	Noise    NoiseSnapshot
	DMC      DMCSnapshot

	SeqCycle      int
	SeqFiveStep   bool
	SeqIrqInhibit bool
	SeqIrqFlag    bool

	CycleParity bool
}

func (a *APU) Snapshot() Snapshot {
	return Snapshot{
		Pulse1: a.pulse1.snapshot(), Pulse2: a.pulse2.snapshot(),
		Triangle: a.triangle.snapshot(), Noise: a.noise.snapshot(), DMC: a.dmc.snapshot(),
		SeqCycle: a.seq.cycle, SeqFiveStep: a.seq.fiveStep,
		SeqIrqInhibit: a.seq.irqInhibit, SeqIrqFlag: a.seq.irqFlag,
		CycleParity: a.cycleParity,
	}
}

func (a *APU) Restore(s Snapshot) {
	a.pulse1.restore(s.Pulse1)
	a.pulse2.restore(s.Pulse2)
	a.triangle.restore(s.Triangle)
	a.noise.restore(s.Noise)
	a.dmc.restore(s.DMC)
	a.seq.cycle, a.seq.fiveStep = s.SeqCycle, s.SeqFiveStep
	a.seq.irqInhibit, a.seq.irqFlag = s.SeqIrqInhibit, s.SeqIrqFlag
	a.cycleParity = s.CycleParity
}
