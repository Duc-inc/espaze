package audio

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
)

func TestRAMReadWriteRoundTrip(t *testing.T) {
	b := New(psg.New())
	b.Write(0x10, 0x99)
	if v := b.Read(0x10); v != 0x99 {
		t.Fatalf("Read(0x10) = %#02x, want 0x99", v)
	}
}

func TestPSGPortWriteReachesTheChip(t *testing.T) {
	b := New(psg.New())
	b.Out(0x00, 0x9F) // latch+attenuate channel 0 to silence
	b.sound.Step(1000)
	for _, s := range b.sound.DrainSamples() {
		if s != 0 {
			t.Fatalf("expected silence after attenuating channel 0 fully, got %d", s)
		}
	}
}
