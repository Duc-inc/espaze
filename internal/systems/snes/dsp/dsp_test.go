package dsp

import "testing"

func TestChannelProducesVaryingOutputWhenKeyedOn(t *testing.T) {
	d := New()
	d.WriteVolume(0, 100)
	d.WritePitchLow(0, 0x00)
	d.WritePitchHigh(0, 0x10)
	for i := 0; i < 32; i++ {
		d.WriteWaveByte(0, byte(i*8-128))
	}
	d.KeyOn(0)

	d.Step(20000)
	samples := d.DrainSamples()
	if len(samples) == 0 {
		t.Fatal("expected samples once a channel is keyed on")
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

func TestKeyOffSilencesChannel(t *testing.T) {
	d := New()
	d.WriteVolume(0, 100)
	d.WritePitchHigh(0, 0x10)
	d.WriteWaveByte(0, 100)
	d.KeyOn(0)
	d.KeyOff(0)

	d.Step(5000)
	for _, s := range d.DrainSamples() {
		if s != 0 {
			t.Fatalf("expected silence after key-off, got %d", s)
		}
	}
}

func TestUnkeyedChannelIsSilent(t *testing.T) {
	d := New()
	d.Step(1000)
	for _, s := range d.DrainSamples() {
		if s != 0 {
			t.Fatalf("expected silence with no channel keyed on, got %d", s)
		}
	}
}
