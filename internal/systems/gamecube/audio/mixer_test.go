package audio

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/adpcm"
)

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

func TestADPCMChannelDecodesRealCompressedData(t *testing.T) {
	m := New()
	var coefs adpcm.Coefficients // zero coefficients: prediction is always 0

	// One frame: predictor 0, scale 0, all nibbles = 1 -> every
	// decoded sample is exactly 1.
	frame := []byte{0x00, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}
	m.SetADPCMChannel(0, coefs, frame, 255, true)

	m.Step(14) // exactly one frame's worth of samples
	samples := m.DrainSamples()
	if len(samples) != 14 {
		t.Fatalf("got %d samples, want 14", len(samples))
	}
	for i, s := range samples {
		if s != 1 {
			t.Fatalf("samples[%d] = %d, want 1 (decoded from real ADPCM data, not silence)", i, s)
		}
	}
}

func TestADPCMChannelGoesSilentAfterDataExhausted(t *testing.T) {
	m := New()
	var coefs adpcm.Coefficients
	frame := []byte{0x00, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11} // one frame, 14 samples
	m.SetADPCMChannel(0, coefs, frame, 255, true)

	m.Step(20) // more than the one frame provides
	samples := m.DrainSamples()
	if samples[19] != 0 {
		t.Fatalf("samples[19] = %d, want 0 (stream exhausted)", samples[19])
	}
}
