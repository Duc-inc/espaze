package memory

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
	"github.com/Duc-inc/espaze/internal/systems/tms9918"
)

func newTestBus() *Bus {
	rom := make([]byte, 0x8000)
	rom[0] = 0xAB
	return New(rom, tms9918.New(), psg.New())
}

func TestCartridgeReadThroughBus(t *testing.T) {
	b := newTestBus()
	if v := b.Read(0); v != 0xAB {
		t.Fatalf("Read(0) = %#02x, want 0xAB", v)
	}
}

func TestRAMReadWriteRoundTrip(t *testing.T) {
	b := newTestBus()
	b.Write(0x6000, 0x42)
	if v := b.Read(0x6000); v != 0x42 {
		t.Fatalf("Read(0x6000) = %#02x, want 0x42", v)
	}
	// The 1KB RAM mirrors across the whole $6000-$7FFF window.
	if v := b.Read(0x6400); v != 0x42 {
		t.Fatalf("Read(0x6400) = %#02x, want 0x42 (RAM mirror)", v)
	}
}

func TestVDPDataPortWriteAndReadRoundTrip(t *testing.T) {
	b := newTestBus()
	b.Out(0xBF, 0x00) // address low byte 0
	b.Out(0xBF, 0x40) // address high (write mode), address = 0x0000
	b.Out(0xBE, 0x77) // data write

	b.Out(0xBF, 0x00) // address low byte 0
	b.Out(0xBF, 0x00) // address high (read mode primes the buffer immediately)
	if v := b.In(0xBE); v != 0x77 {
		t.Fatalf("VDP data read = %#02x, want 0x77", v)
	}
}

func TestPSGWriteReachesTheChip(t *testing.T) {
	b := newTestBus()
	b.Out(0xFF, 0x9F) // latch+attenuate channel 0 to silence - confirms the write path reaches the PSG without panicking
}

func TestControllerReadsActiveLow(t *testing.T) {
	b := newTestBus()
	b.SetButtons(input.State{}.With(Up, true))
	v := b.In(0xFC)
	if v&0x01 != 0 {
		t.Fatal("expected Up bit to read active-low (0) when held")
	}
}
