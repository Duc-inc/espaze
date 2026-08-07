package psg

type channelSnapshot struct {
	Freq                      uint16
	Enabled, DDAMode          bool
	Volume, Pan               byte
	Wave                      [32]byte
	WaveWriteIndex, PlayIndex byte
	PhaseAccum                int
	NoiseMode                 bool
	NoiseFreq                 byte
	NoiseLFSR                 uint32
	NoiseCounter              int
	NoiseOutput               bool
}

func snapshotChannel(c *channel) channelSnapshot {
	return channelSnapshot{
		Freq: c.freq, Enabled: c.enabled, DDAMode: c.ddaMode, Volume: c.volume, Pan: c.pan,
		Wave: c.wave, WaveWriteIndex: c.waveWriteIndex, PlayIndex: c.playIndex, PhaseAccum: c.phaseAccum,
		NoiseMode: c.noiseMode, NoiseFreq: c.noiseFreq, NoiseLFSR: c.noiseLFSR,
		NoiseCounter: c.noiseCounter, NoiseOutput: c.noiseOutput,
	}
}

func restoreChannel(c *channel, s channelSnapshot) {
	c.freq, c.enabled, c.ddaMode, c.volume, c.pan = s.Freq, s.Enabled, s.DDAMode, s.Volume, s.Pan
	c.wave, c.waveWriteIndex, c.playIndex, c.phaseAccum = s.Wave, s.WaveWriteIndex, s.PlayIndex, s.PhaseAccum
	c.noiseMode, c.noiseFreq, c.noiseLFSR = s.NoiseMode, s.NoiseFreq, s.NoiseLFSR
	c.noiseCounter, c.noiseOutput = s.NoiseCounter, s.NoiseOutput
}

// Snapshot captures the whole PSG's state.
type Snapshot struct {
	Selected           byte
	Channels           [6]channelSnapshot
	MainVolL, MainVolR byte
	SampleCycles       float64
}

// Snapshot captures the PSG's current state.
func (p *PSG) Snapshot() Snapshot {
	var chans [6]channelSnapshot
	for i := range p.channels {
		chans[i] = snapshotChannel(&p.channels[i])
	}
	return Snapshot{
		Selected: p.selected, Channels: chans,
		MainVolL: p.mainVolL, MainVolR: p.mainVolR, SampleCycles: p.sampleCycles,
	}
}

// Restore reinstates a previously captured Snapshot.
func (p *PSG) Restore(s Snapshot) {
	p.selected = s.Selected
	for i := range p.channels {
		restoreChannel(&p.channels[i], s.Channels[i])
	}
	p.mainVolL, p.mainVolR, p.sampleCycles = s.MainVolL, s.MainVolR, s.SampleCycles
}
