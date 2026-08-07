package psg

type toneSnapshot struct {
	Freq, Counter uint16
	Output        bool
	Atten         byte
}

func (c *toneChannel) snapshot() toneSnapshot {
	return toneSnapshot{Freq: c.freq, Counter: c.counter, Output: c.output, Atten: c.atten}
}

func (c *toneChannel) restore(s toneSnapshot) {
	c.freq, c.counter, c.output, c.atten = s.Freq, s.Counter, s.Output, s.Atten
}

type noiseSnapshot struct {
	ShiftRate       byte
	FBMode          bool
	LFSR, Counter   uint16
	Atten           byte
	Tone2PrevOutput bool
}

func (c *noiseChannel) snapshot() noiseSnapshot {
	return noiseSnapshot{
		ShiftRate: c.shiftRate, FBMode: c.fbMode, LFSR: c.lfsr, Counter: c.counter,
		Atten: c.atten, Tone2PrevOutput: c.tone2PrevOutput,
	}
}

func (c *noiseChannel) restore(s noiseSnapshot) {
	c.shiftRate, c.fbMode, c.lfsr, c.counter = s.ShiftRate, s.FBMode, s.LFSR, s.Counter
	c.atten, c.tone2PrevOutput = s.Atten, s.Tone2PrevOutput
}

// Snapshot captures the whole PSG's state.
type Snapshot struct {
	Tone0, Tone1, Tone2 toneSnapshot
	Noise               noiseSnapshot
	LatchedChannel      byte
	LatchedType         byte
	Prescaler           int
}

func (p *PSG) Snapshot() Snapshot {
	return Snapshot{
		Tone0: p.tone0.snapshot(), Tone1: p.tone1.snapshot(), Tone2: p.tone2.snapshot(),
		Noise:          p.noise.snapshot(),
		LatchedChannel: p.latchedChannel, LatchedType: p.latchedType,
		Prescaler: p.prescaler,
	}
}

func (p *PSG) Restore(s Snapshot) {
	p.tone0.restore(s.Tone0)
	p.tone1.restore(s.Tone1)
	p.tone2.restore(s.Tone2)
	p.noise.restore(s.Noise)
	p.latchedChannel, p.latchedType = s.LatchedChannel, s.LatchedType
	p.prescaler = s.Prescaler
}
