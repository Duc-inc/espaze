package apu

import "testing"

type testMemory struct {
	data [65536]byte
}

func (m *testMemory) Read(addr uint16) byte { return m.data[addr] }

func TestPulseChannelProducesVaryingWaveform(t *testing.T) {
	a := New(&testMemory{})
	a.WriteRegister(0x4015, 0x01) // enable pulse 1
	a.WriteRegister(0x4000, 0xBF) // duty=50%, constant volume 15, halt/loop
	a.WriteRegister(0x4001, 0x00) // sweep off
	a.WriteRegister(0x4002, 0x00) // timer low
	a.WriteRegister(0x4003, 0x09) // timer high=1 (period 256, audible), length load + trigger

	a.Step(29830) // just over one full frame-sequencer cycle
	samples := a.DrainSamples()

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

func TestLengthCounterSilencesChannel(t *testing.T) {
	a := New(&testMemory{})
	a.WriteRegister(0x4015, 0x01)
	a.WriteRegister(0x4000, 0x1F) // duty=00, NOT halted, constant volume 15
	a.WriteRegister(0x4002, 0x00)
	a.WriteRegister(0x4003, 0x08) // length index 1 -> table value 254, trigger

	if !a.pulse1.active() {
		t.Fatal("channel should be active right after trigger")
	}

	// One length tick happens every half-frame (~14913 cycles for the
	// first one); step past several to force the count down to zero.
	for i := 0; i < 260; i++ {
		a.Step(14913)
	}

	if a.pulse1.active() {
		t.Fatal("channel should have been silenced once its length counter reached zero")
	}
}

func TestFrameSequencerIRQFiresInFourStepMode(t *testing.T) {
	a := New(&testMemory{})
	a.WriteRegister(0x4017, 0x00) // 4-step mode, IRQ allowed

	a.Step(seqStep4 + 1)
	if !a.IRQPending() {
		t.Fatal("expected the frame IRQ to fire at the end of 4-step mode")
	}

	v := a.ReadRegister(0x4015)
	if v&0x40 == 0 {
		t.Fatal("expected $4015 to report the frame IRQ flag")
	}
	if a.IRQPending() {
		t.Fatal("reading $4015 should have cleared the frame IRQ")
	}
}

func TestFrameSequencerNoIRQInFiveStepMode(t *testing.T) {
	a := New(&testMemory{})
	a.WriteRegister(0x4017, 0x80) // 5-step mode

	a.Step(seqStep5 + 1)
	if a.IRQPending() {
		t.Fatal("5-step mode should never raise the frame IRQ")
	}
}

func TestDMCPlaysBackSampleBytes(t *testing.T) {
	mem := &testMemory{}
	mem.data[0xC000] = 0xFF // all bits set -> level should climb
	a := New(mem)
	a.dmc.mem = mem

	a.WriteRegister(0x4010, 0x0F) // fastest rate, no loop, no IRQ
	a.WriteRegister(0x4012, 0x00) // sample address $C000
	a.WriteRegister(0x4013, 0x00) // sample length 1 byte
	a.WriteRegister(0x4011, 0x00) // start level at 0
	a.WriteRegister(0x4015, 0x10) // enable DMC

	if !a.dmc.active() {
		t.Fatal("DMC should be active right after being enabled with bytesLeft > 0")
	}

	a.Step(int(dmcRateTable[0])*8 + 10) // enough ticks to consume all 8 bits
	if a.dmc.output() == 0 {
		t.Fatal("expected DMC output level to have climbed from playing back 0xFF")
	}
}

func TestNoiseChannelSilentWithoutTrigger(t *testing.T) {
	a := New(&testMemory{})
	a.Step(10000)
	samples := a.DrainSamples()
	for _, s := range samples {
		if s != 0 {
			t.Fatalf("expected silence with nothing enabled, got sample %d", s)
		}
	}
}
