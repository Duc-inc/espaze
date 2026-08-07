package apu

import "testing"

func TestFIFOAPlaybackProducesVaryingOutput(t *testing.T) {
	a := New()
	a.WriteSoundCntH(0x0304) // A: full volume, enabled both L/R

	for i := 0; i < 64; i++ {
		a.WriteFIFOA(byte(i*4 - 128))
	}

	a.Step(100000)
	samples := a.DrainSamples()
	if len(samples) == 0 {
		t.Fatal("expected samples once channel A is enabled and fed")
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
		t.Fatalf("expected varying output as the FIFO drains, got constant %d", min)
	}
}

func TestDisabledChannelIsSilent(t *testing.T) {
	a := New()
	a.WriteFIFOA(100)
	a.Step(10000)
	for _, s := range a.DrainSamples() {
		if s != 0 {
			t.Fatalf("expected silence with channel A not enabled, got %d", s)
		}
	}
}

func TestResetFIFOClearsQueuedSamples(t *testing.T) {
	a := New()
	a.WriteFIFOA(1)
	a.WriteFIFOA(2)
	a.WriteSoundCntH(0x0800) // reset FIFO A bit
	if len(a.fifoA) != 0 {
		t.Fatalf("fifoA length = %d, want 0 after reset", len(a.fifoA))
	}
}
