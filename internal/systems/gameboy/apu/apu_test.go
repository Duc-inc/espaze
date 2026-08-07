package apu

import "testing"

func TestSilentWhenPoweredOff(t *testing.T) {
	a := New()
	a.Step(9510)
	samples := a.DrainSamples()
	for _, s := range samples {
		if s != 0 {
			t.Fatalf("expected silence while powered off, got sample %d", s)
		}
	}
}

func TestSquareChannelProducesSound(t *testing.T) {
	a := New()
	a.WriteRegister(0xFF26, 0x80) // power on
	a.WriteRegister(0xFF24, 0x77) // NR50: master volume maxed on both sides

	// Channel 2: 50% duty, static volume 15 (envelope period 0 = no fade),
	// a mid-range frequency, then trigger.
	a.WriteRegister(0xFF16, 0x80) // NR21: duty=10(50%), length=0
	a.WriteRegister(0xFF17, 0xF0) // NR22: volume=15, no envelope sweep
	a.WriteRegister(0xFF18, 0x00) // NR23: frequency low byte
	a.WriteRegister(0xFF19, 0x87) // NR24: frequency high=3, trigger=1

	// Step one full frame's worth of cycles (matches gameboy.StepFrame).
	a.Step(154 * 456)
	samples := a.DrainSamples()

	if len(samples) == 0 {
		t.Fatal("expected samples to be generated once a channel is playing")
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
		t.Fatalf("expected a varying waveform (square wave), got a constant %d", min)
	}
	if max == 0 && min == 0 {
		t.Fatal("expected non-silent output from an active square channel")
	}
}

func TestLengthCounterSilencesChannel(t *testing.T) {
	a := New()
	a.WriteRegister(0xFF26, 0x80)

	a.WriteRegister(0xFF16, 0x3F) // NR21: duty=00, length=63 (near-max, 1 tick from silence)
	a.WriteRegister(0xFF17, 0xF0) // full volume, no envelope
	a.WriteRegister(0xFF19, 0xC0) // NR24: length enable=1, trigger=1

	if !a.ch2.enabled {
		t.Fatal("channel should be enabled right after trigger")
	}

	// One length tick happens every 2 frame-sequencer steps (256Hz): step
	// the sequencer past a handful of 512Hz ticks to force one through.
	a.Step(sequencerPeriod * 2)

	if a.ch2.enabled {
		t.Fatal("channel should have been silenced once its length counter reached zero")
	}
}

func TestSaveStateRoundTripPreservesChannelState(t *testing.T) {
	a := New()
	a.WriteRegister(0xFF26, 0x80)
	a.WriteRegister(0xFF16, 0x80)
	a.WriteRegister(0xFF17, 0xF0)
	a.WriteRegister(0xFF19, 0x87)
	a.Step(1000)

	snap := a.Snapshot()

	fresh := New()
	fresh.Restore(snap)

	if fresh.ch2.frequency != a.ch2.frequency || fresh.ch2.enabled != a.ch2.enabled {
		t.Fatalf("restored channel state doesn't match: got freq=%d enabled=%v, want freq=%d enabled=%v",
			fresh.ch2.frequency, fresh.ch2.enabled, a.ch2.frequency, a.ch2.enabled)
	}
}
