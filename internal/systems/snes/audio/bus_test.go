package audio

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/snes/dsp"
)

func TestARAMReadWriteRoundTrip(t *testing.T) {
	b := New(dsp.New(), &Ports{})
	b.Write8(0x10, 0x99)
	if v := b.Read8(0x10); v != 0x99 {
		t.Fatalf("Read8(0x10) = %#02x, want 0x99", v)
	}
}

func TestSharedPortsVisibleThroughBus(t *testing.T) {
	ports := &Ports{}
	b := New(dsp.New(), ports)
	b.Write8(0xF4, 0x42)
	if v := ports.Read(0); v != 0x42 {
		t.Fatalf("port 0 = %#02x, want 0x42", v)
	}
	if v := b.Read8(0xF4); v != 0x42 {
		t.Fatalf("Read8(0xF4) = %#02x, want 0x42", v)
	}
}

func TestDSPRegisterWriteReachesTheChip(t *testing.T) {
	sound := dsp.New()
	b := New(sound, &Ports{})

	b.Write8(0xF2, 0x00) // select channel 0's volume register
	b.Write8(0xF3, 100)
	b.Write8(0xF2, 0x03) // channel 0 pitch high
	b.Write8(0xF3, 0x10)
	b.Write8(0xF2, 0x04) // channel 0 wave data
	for i := 0; i < 32; i++ {
		b.Write8(0xF3, byte(i*8-128))
	}
	b.Write8(0xF2, 0x4C) // KON
	b.Write8(0xF3, 0x01) // key on channel 0

	sound.Step(20000)
	samples := sound.DrainSamples()
	if len(samples) == 0 {
		t.Fatal("expected samples once keyed on through the DSP register port")
	}
}
