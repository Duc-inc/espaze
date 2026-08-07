package memory

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/atari2600/riot"
	"github.com/Duc-inc/espaze/internal/systems/atari2600/tia"
)

func newTestBus() *Bus {
	rom := make([]byte, 0x1000)
	rom[0], rom[1] = 0xAB, 0xCD
	return New(rom, tia.New(), riot.New())
}

func TestCartridgeReadMirrorsIntoUpperWindow(t *testing.T) {
	b := newTestBus()
	if v := b.Read(0x1000); v != 0xAB {
		t.Fatalf("Read(0x1000) = %#02x, want 0xAB", v)
	}
	if v := b.Read(0xF000); v != 0xAB { // A12 set is the only bit that matters
		t.Fatalf("Read(0xF000) = %#02x, want 0xAB (mirrored)", v)
	}
}

func TestRIOTRAMReadWriteThroughBus(t *testing.T) {
	b := newTestBus()
	b.Write(0x0080, 0x77)
	if v := b.Read(0x0080); v != 0x77 {
		t.Fatalf("Read(0x0080) = %#02x, want 0x77", v)
	}
}

func TestTIARegisterWriteThroughBus(t *testing.T) {
	b := newTestBus()
	b.Write(0x09, 0x1E) // COLUBK
	// No direct getter on tia.TIA for COLUBK, so just confirm the write
	// didn't panic and land somewhere sane by reading a collision
	// register (must still be zero, proving the write didn't corrupt
	// unrelated TIA state).
	if v := b.Read(0x00); v != 0 {
		t.Fatalf("CXM0P after an unrelated write = %#02x, want 0", v)
	}
}

func TestTimerWriteThroughBus(t *testing.T) {
	b := newTestBus()
	b.Write(0x0294, 10)               // TIM1T
	if v := b.Read(0x0284); v != 10 { // INTIM
		t.Fatalf("INTIM after TIM1T write = %d, want 10", v)
	}
}
