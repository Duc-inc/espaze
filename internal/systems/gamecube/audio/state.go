package audio

type channelSnapshot struct {
	Sample  int16
	Volume  byte
	Enabled bool
}

// Snapshot captures every channel's current state.
type Snapshot struct {
	Channels [channelCount]channelSnapshot
}

// Snapshot captures the mixer's current state.
func (m *Mixer) Snapshot() Snapshot {
	var s Snapshot
	for i, ch := range m.channels {
		s.Channels[i] = channelSnapshot{Sample: ch.sample, Volume: ch.volume, Enabled: ch.enabled}
	}
	return s
}

// Restore reinstates a previously captured Snapshot.
func (m *Mixer) Restore(s Snapshot) {
	for i, cs := range s.Channels {
		m.channels[i] = channel{sample: cs.Sample, volume: cs.Volume, enabled: cs.Enabled}
	}
}
