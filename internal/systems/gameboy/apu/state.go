package apu

// Snapshot captures every piece of APU state a save state needs to
// resume audio exactly where it left off.
type Snapshot struct {
	Ch1 SquareSnapshot
	Ch2 SquareSnapshot
	Ch3 WaveSnapshot
	Ch4 NoiseSnapshot

	SequencerTimer int
	SequencerStep  int

	MasterLeft, MasterRight, Panning byte
	PowerOn                          bool
}

type SquareSnapshot struct {
	Duty, DutyStep  byte
	Frequency       uint16
	Timer           int
	Enabled         bool
	LengthCounter   int
	LengthEnabled   bool
	EnvelopeRaw     byte
	EnvelopeVolume  byte
	EnvelopeTimer   byte
	SweepShadowFreq uint16
	SweepTimer      byte
	SweepEnabled    bool
}

type WaveSnapshot struct {
	DacEnabled    bool
	VolumeCode    byte
	Frequency     uint16
	Timer         int
	Position      byte
	RAM           [16]byte
	Enabled       bool
	LengthCounter int
	LengthEnabled bool
}

type NoiseSnapshot struct {
	ShiftClockFreq, DivisorCode byte
	WidthMode                   bool
	LFSR                        uint16
	Timer                       int
	Enabled                     bool
	LengthCounter               int
	LengthEnabled               bool
	EnvelopeRaw                 byte
	EnvelopeVolume              byte
	EnvelopeTimer               byte
}

func (a *APU) Snapshot() Snapshot {
	return Snapshot{
		Ch1:            snapshotSquare(a.ch1),
		Ch2:            snapshotSquare(a.ch2),
		Ch3:            snapshotWave(a.ch3),
		Ch4:            snapshotNoise(a.ch4),
		SequencerTimer: a.seq.timer,
		SequencerStep:  a.seq.step,
		MasterLeft:     a.masterLeft,
		MasterRight:    a.masterRight,
		Panning:        a.panning,
		PowerOn:        a.powerOn,
	}
}

func (a *APU) Restore(s Snapshot) {
	restoreSquare(a.ch1, s.Ch1)
	restoreSquare(a.ch2, s.Ch2)
	restoreWave(a.ch3, s.Ch3)
	restoreNoise(a.ch4, s.Ch4)
	a.seq.timer = s.SequencerTimer
	a.seq.step = s.SequencerStep
	a.masterLeft, a.masterRight, a.panning = s.MasterLeft, s.MasterRight, s.Panning
	a.powerOn = s.PowerOn
}

func snapshotSquare(c *squareChannel) SquareSnapshot {
	return SquareSnapshot{
		Duty: c.duty, DutyStep: c.dutyStep, Frequency: c.frequency,
		Timer: c.timer, Enabled: c.enabled,
		LengthCounter: c.length.counter, LengthEnabled: c.length.enabled,
		EnvelopeRaw: c.envelope.raw, EnvelopeVolume: c.envelope.volume, EnvelopeTimer: c.envelope.timer,
		SweepShadowFreq: c.sweep.shadowFreq, SweepTimer: c.sweep.timer, SweepEnabled: c.sweep.enabled,
	}
}

func restoreSquare(c *squareChannel, s SquareSnapshot) {
	c.duty, c.dutyStep, c.frequency = s.Duty, s.DutyStep, s.Frequency
	c.timer, c.enabled = s.Timer, s.Enabled
	c.length.counter, c.length.enabled = s.LengthCounter, s.LengthEnabled
	c.envelope.writeRegister(s.EnvelopeRaw)
	c.envelope.volume, c.envelope.timer = s.EnvelopeVolume, s.EnvelopeTimer
	c.sweep.shadowFreq, c.sweep.timer, c.sweep.enabled = s.SweepShadowFreq, s.SweepTimer, s.SweepEnabled
}

func snapshotWave(c *waveChannel) WaveSnapshot {
	return WaveSnapshot{
		DacEnabled: c.dacEnabled, VolumeCode: c.volumeCode, Frequency: c.frequency,
		Timer: c.timer, Position: c.position, RAM: c.ram, Enabled: c.enabled,
		LengthCounter: c.length.counter, LengthEnabled: c.length.enabled,
	}
}

func restoreWave(c *waveChannel, s WaveSnapshot) {
	c.dacEnabled, c.volumeCode, c.frequency = s.DacEnabled, s.VolumeCode, s.Frequency
	c.timer, c.position, c.ram, c.enabled = s.Timer, s.Position, s.RAM, s.Enabled
	c.length.counter, c.length.enabled = s.LengthCounter, s.LengthEnabled
}

func snapshotNoise(c *noiseChannel) NoiseSnapshot {
	return NoiseSnapshot{
		ShiftClockFreq: c.shiftClockFreq, DivisorCode: c.divisorCode, WidthMode: c.widthMode,
		LFSR: c.lfsr, Timer: c.timer, Enabled: c.enabled,
		LengthCounter: c.length.counter, LengthEnabled: c.length.enabled,
		EnvelopeRaw: c.envelope.raw, EnvelopeVolume: c.envelope.volume, EnvelopeTimer: c.envelope.timer,
	}
}

func restoreNoise(c *noiseChannel, s NoiseSnapshot) {
	c.shiftClockFreq, c.divisorCode, c.widthMode = s.ShiftClockFreq, s.DivisorCode, s.WidthMode
	c.lfsr, c.timer, c.enabled = s.LFSR, s.Timer, s.Enabled
	c.length.counter, c.length.enabled = s.LengthCounter, s.LengthEnabled
	c.envelope.writeRegister(s.EnvelopeRaw)
	c.envelope.volume, c.envelope.timer = s.EnvelopeVolume, s.EnvelopeTimer
}
