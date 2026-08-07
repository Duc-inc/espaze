package ym2612

import "testing"

func TestChannelProducesVaryingWaveformAfterKeyOn(t *testing.T) {
	y := New()

	y.WriteAddress1(0xB0) // channel 0: algorithm/feedback
	y.WriteData1(0x07)    // algorithm 7 (fully parallel), no feedback

	y.WriteAddress1(0x30) // channel 0, operator 1: MUL/DET
	y.WriteData1(0x01)    // multiplier 1

	y.WriteAddress1(0x40) // operator 1 total level (loudest)
	y.WriteData1(0x00)
	y.WriteAddress1(0x50) // operator 1 attack rate (fast)
	y.WriteData1(0x1F)

	y.WriteAddress1(0xA0) // channel 0 F-NUM low
	y.WriteData1(0x00)
	y.WriteAddress1(0xA4) // channel 0 F-NUM high + block
	y.WriteData1(0x22)    // a mid-range frequency

	y.WriteAddress1(0xB4) // channel 0 pan: both channels on
	y.WriteData1(0xC0)

	y.WriteAddress1(0x28) // key on: channel 0, all 4 operators
	y.WriteData1(0xF0)

	y.Step(20000)
	samples := y.DrainSamples()
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

func TestSilentWithoutAnyPanEnabled(t *testing.T) {
	y := New()
	// Key on channel 0 but never enable its L/R pan bits - real
	// hardware outputs nothing for a channel with both disabled.
	y.WriteAddress1(0x40)
	y.WriteData1(0x00)
	y.WriteAddress1(0x28)
	y.WriteData1(0xF0)

	y.Step(20000)
	for _, s := range y.DrainSamples() {
		if s != 0 {
			t.Fatalf("expected silence with pan disabled, got sample %d", s)
		}
	}
}

func TestKeyOnOffTargetsCorrectPart(t *testing.T) {
	y := New()
	// Key on channel 4 (part 2, channel index 1 within that part):
	// bit2 set selects part 2, bits0-1=01 select channel index 1.
	y.writeKeyOnOff(0x05 | 0xF0)
	if !y.channels[4].ops[0].keyOn {
		t.Fatal("expected channel 4's operators to be keyed on")
	}
	if y.channels[1].ops[0].keyOn {
		t.Fatal("channel 1 (part 1) should be unaffected")
	}
}

func TestFNumHighAndBlockDecode(t *testing.T) {
	y := New()
	y.WriteAddress1(0xA0)
	y.WriteData1(0x34)
	y.WriteAddress1(0xA4)
	y.WriteData1(0x22) // block=4 (bits5-3=100), F-NUM high bits=010

	ch := &y.channels[0]
	wantFNum := uint16(0x02)<<8 | 0x34
	if ch.fnum != wantFNum {
		t.Fatalf("fnum = %#04x, want %#04x", ch.fnum, wantFNum)
	}
	if ch.block != 4 {
		t.Fatalf("block = %d, want 4", ch.block)
	}
}
