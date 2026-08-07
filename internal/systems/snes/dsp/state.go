package dsp

type channelSnapshot struct {
	Enabled               bool
	Volume                byte
	Pitch                 uint16
	Wave                  [32]int8
	WaveWriteIdx, PlayIdx byte
	PhaseAccum            int
}

func snapshotChannel(c *channel) channelSnapshot {
	return channelSnapshot{
		Enabled: c.enabled, Volume: c.volume, Pitch: c.pitch,
		Wave: c.wave, WaveWriteIdx: c.waveWriteIdx, PlayIdx: c.playIdx, PhaseAccum: c.phaseAccum,
	}
}

func restoreChannel(c *channel, s channelSnapshot) {
	c.enabled, c.volume, c.pitch = s.Enabled, s.Volume, s.Pitch
	c.wave, c.waveWriteIdx, c.playIdx, c.phaseAccum = s.Wave, s.WaveWriteIdx, s.PlayIdx, s.PhaseAccum
}

// Snapshot captures the whole DSP's state.
type Snapshot struct {
	Channels     [8]channelSnapshot
	SampleCycles float64
}

// Snapshot captures the DSP's current state.
func (d *DSP) Snapshot() Snapshot {
	var chans [8]channelSnapshot
	for i := range d.channels {
		chans[i] = snapshotChannel(&d.channels[i])
	}
	return Snapshot{Channels: chans, SampleCycles: d.sampleCycles}
}

// Restore reinstates a previously captured Snapshot.
func (d *DSP) Restore(s Snapshot) {
	for i := range d.channels {
		restoreChannel(&d.channels[i], s.Channels[i])
	}
	d.sampleCycles = s.SampleCycles
}
