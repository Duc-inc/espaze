package audio

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/genesis/ym2612"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
)

func TestRAMReadWriteMirrorsAcrossWindow(t *testing.T) {
	b := New(ym2612.New(), psg.New())
	b.Write(0x0010, 0x99)
	if v := b.Read(0x2010); v != 0x99 { // $2000-$3FFF mirrors $0000-$1FFF
		t.Fatalf("mirrored read = %#02x, want 0x99", v)
	}
}

func TestYM2612PortWritesReachTheChip(t *testing.T) {
	ym := ym2612.New()
	b := New(ym, psg.New())

	b.Write(0x4000, 0xB0) // address 1: channel 0 algorithm/feedback register
	b.Write(0x4001, 0x07) // algorithm 7 (fully parallel), no feedback
	b.Write(0x4000, 0xB4) // address 1: channel 0 pan register
	b.Write(0x4001, 0xC0) // data 1: both L/R on
	b.Write(0x4000, 0x40) // operator 1 total level (loudest)
	b.Write(0x4001, 0x00)
	b.Write(0x4000, 0x50) // operator 1 attack rate (fast)
	b.Write(0x4001, 0x1F)
	b.Write(0x4000, 0xA0) // channel 0 F-NUM low
	b.Write(0x4001, 0x00)
	b.Write(0x4000, 0xA4) // channel 0 F-NUM high + block
	b.Write(0x4001, 0x22) // a mid-range frequency
	b.Write(0x4000, 0x28) // key on register
	b.Write(0x4001, 0xF0) // key on channel 0, all operators

	ym.Step(2000)
	samples := ym.DrainSamples()
	if len(samples) == 0 {
		t.Fatal("expected samples once keyed on through the Z80 bus")
	}
	allZero := true
	for _, s := range samples {
		if s != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Fatal("expected non-silent output after keying on channel 0 through the Z80 bus")
	}
}

func TestPSGPortWriteReachesTheChip(t *testing.T) {
	b := New(ym2612.New(), psg.New())
	b.Write(0x7F11, 0x9F) // latch+attenuate channel 0 to silence
	b.sound.Step(1000)
	for _, s := range b.sound.DrainSamples() {
		if s != 0 {
			t.Fatalf("expected silence after attenuating channel 0 fully, got %d", s)
		}
	}
}

func TestBusRequestHaltsAndAcknowledges(t *testing.T) {
	b := New(ym2612.New(), psg.New())
	if b.BusAcknowledged() {
		t.Fatal("should not be acknowledged before a request")
	}
	b.RequestBus(true)
	if !b.BusAcknowledged() || !b.Halted() {
		t.Fatal("expected bus request to be acknowledged and halt the coprocessor")
	}
}
