package psg

import "testing"

func TestToneChannelProducesVaryingWaveform(t *testing.T) {
	p := New()
	p.Write(0x80) // latch: channel 0, frequency, low nibble = 0
	p.Write(0x02) // freq low = 2
	p.Write(0x00) // data: freq high = 0 -> freq = 2 (fast, audible tone)
	p.Write(0x90) // latch: channel 0, attenuation = 0 (loudest)

	p.Step(100000)
	samples := p.DrainSamples()

	if len(samples) == 0 {
		t.Fatal("expected samples once a channel is playing")
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

func TestSilentWhenFullyAttenuated(t *testing.T) {
	p := New() // every channel starts at max attenuation (silent)
	p.Write(0x80)
	p.Write(0x02)
	p.Write(0x00)
	// attenuation left at its default (0x0F = silent)

	p.Step(100000)
	for _, s := range p.DrainSamples() {
		if s != 0 {
			t.Fatalf("expected silence with max attenuation, got sample %d", s)
		}
	}
}

func TestFrequencyLatchThenDataByte(t *testing.T) {
	p := New()
	p.Write(0x81) // latch: channel 0, frequency, low nibble = 1
	p.Write(0x2A) // data byte: high 6 bits = 0x2A

	want := uint16(0x2A)<<4 | 1
	if p.tone0.freq != want {
		t.Fatalf("tone0.freq = %#04x, want %#04x", p.tone0.freq, want)
	}
}

func TestAttenuationWrite(t *testing.T) {
	p := New()
	// 1 01 1 1010: latch, channel 1 (bits6-5=01), attenuation (bit4=1), value 0x0A.
	p.Write(0xBA)
	if p.tone1.atten != 0x0A {
		t.Fatalf("tone1.atten = %#02x, want 0x0A", p.tone1.atten)
	}
}

func TestNoiseControlResetsLFSR(t *testing.T) {
	p := New()
	p.noise.lfsr = 0x0001 // perturb it away from the reset value
	p.Write(0xE4)         // latch: channel 3, control, white noise mode

	if p.noise.lfsr != 0x8000 {
		t.Fatalf("lfsr after control write = %#04x, want 0x8000 (reset)", p.noise.lfsr)
	}
	if !p.noise.fbMode {
		t.Fatal("expected white noise mode (bit 2 set)")
	}
}

func TestNoiseShiftRate3ClockedByTone2(t *testing.T) {
	p := New()
	p.Write(0xE7) // latch: channel 3, control, shift rate 3 (clocked by tone2)
	p.Write(0x84) // latch: channel 2, frequency, low nibble = 4 (short period, ticks often)
	p.Write(0x00)

	before := p.noise.lfsr
	p.Step(1000)
	if p.noise.lfsr == before {
		t.Fatal("expected the LFSR to have clocked at least once via tone2's transitions")
	}
}
