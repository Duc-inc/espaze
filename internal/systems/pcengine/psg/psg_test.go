package psg

import "testing"

func TestChannelProducesVaryingOutputWhenEnabled(t *testing.T) {
	p := New()
	p.SelectChannel(0)
	p.WriteFreqLow(0x10)
	p.WriteFreqHigh(0x00)
	p.WriteWaveData(0x00) // control still off, but writeWaveData works regardless
	p.WriteControl(0x40)  // DDA mode on: reset write index
	for i := byte(0); i < 32; i++ {
		p.WriteWaveData(i % 16)
	}
	p.WriteControl(0x9F) // enabled, DDA off (plays the table), volume max

	p.Step(20000)
	samples := p.DrainSamples()
	if len(samples) == 0 {
		t.Fatal("expected samples once a channel is enabled")
	}
	min, max := samples[0], samples[0]
	for _, s := range samples {
		if s < min {
			min = s
		}
		if s > max {
			max = s
		}
	}
	if min == max {
		t.Fatalf("expected a varying waveform, got a constant %d", min)
	}
}

func TestDisabledChannelIsSilent(t *testing.T) {
	p := New()
	p.Step(1000)
	for _, s := range p.DrainSamples() {
		if s != 0 {
			t.Fatalf("expected silence with no channel enabled, got %d", s)
		}
	}
}

func TestNoiseModeOnlyAffectsChannels4And5(t *testing.T) {
	p := New()
	p.SelectChannel(0)
	p.WriteNoiseControl(0x9F) // should be ignored on channel 0
	if p.channels[0].noiseMode {
		t.Fatal("channel 0 should not support noise mode")
	}

	p.SelectChannel(4)
	p.WriteNoiseControl(0x9F)
	if !p.channels[4].noiseMode {
		t.Fatal("channel 4 should support noise mode")
	}
}
