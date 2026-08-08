package audio

import "testing"

func TestMixSampleSumsEnabledChannels(t *testing.T) {
	m := New()
	m.SetChannel(0, 100, 255, true)
	m.SetChannel(1, 100, 255, true)
	if v := m.MixSample(); v != 200 {
		t.Fatalf("MixSample = %d, want 200", v)
	}
}

func TestDisabledChannelIsIgnored(t *testing.T) {
	m := New()
	m.SetChannel(0, 32000, 255, false)
	if v := m.MixSample(); v != 0 {
		t.Fatalf("MixSample = %d, want 0 (channel disabled)", v)
	}
}

func TestMixSampleClamps(t *testing.T) {
	m := New()
	for i := 0; i < channelCount; i++ {
		m.SetChannel(i, 32767, 255, true)
	}
	if v := m.MixSample(); v != 32767 {
		t.Fatalf("MixSample = %d, want clamped to 32767", v)
	}
}

func TestStepGeneratesRequestedSampleCount(t *testing.T) {
	m := New()
	m.Step(100)
	if len(m.DrainSamples()) != 100 {
		t.Fatalf("got %d samples, want 100", len(m.DrainSamples()))
	}
}
