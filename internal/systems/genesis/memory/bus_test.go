package memory

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/systems/genesis/vdp"
)

type fakeZ80 struct {
	ram        [0x10000]byte
	busReq     bool
	requestLog []bool
}

func (f *fakeZ80) Read(addr uint16) byte     { return f.ram[addr] }
func (f *fakeZ80) Write(addr uint16, v byte) { f.ram[addr] = v }
func (f *fakeZ80) RequestBus(held bool)      { f.busReq = held; f.requestLog = append(f.requestLog, held) }
func (f *fakeZ80) BusAcknowledged() bool     { return f.busReq }

func newTestBus() (*Bus, *fakeZ80) {
	z80 := &fakeZ80{}
	rom := make([]byte, 0x1000)
	rom[0], rom[1] = 0xAB, 0xCD
	return New(rom, vdp.New(), z80), z80
}

func TestCartridgeReadReturnsROMBytes(t *testing.T) {
	b, _ := newTestBus()
	if v := b.Read8(0); v != 0xAB {
		t.Fatalf("Read8(0) = %#02x, want 0xAB", v)
	}
	if v := b.Read16(0); v != 0xABCD {
		t.Fatalf("Read16(0) = %#04x, want 0xABCD", v)
	}
}

func TestWorkRAMReadWriteRoundTrip(t *testing.T) {
	b, _ := newTestBus()
	b.Write32(0xFF0000, 0xDEADBEEF)
	if v := b.Read32(0xFF0000); v != 0xDEADBEEF {
		t.Fatalf("Read32 = %#08x, want 0xDEADBEEF", v)
	}
}

func TestZ80WindowDelegatesToBridge(t *testing.T) {
	b, z80 := newTestBus()
	b.Write8(0xA01000, 0x42)
	if z80.ram[0x1000] != 0x42 {
		t.Fatalf("Z80 RAM[0x1000] = %#02x, want 0x42", z80.ram[0x1000])
	}
	if v := b.Read8(0xA01000); v != 0x42 {
		t.Fatalf("Read8 through Z80 window = %#02x, want 0x42", v)
	}
}

func TestBusRequestRegisterDelegates(t *testing.T) {
	b, z80 := newTestBus()
	b.Write8(0xA11100, 0x01)
	if !z80.busReq {
		t.Fatal("expected bus request to be forwarded to the Z80 bridge")
	}
}

func TestControllerReportsButtonsByTHLine(t *testing.T) {
	b, _ := newTestBus()
	b.SetButtons(input.State{}.With(Up, true).With(Start, true))

	b.Write8(0xA10003, 0x00) // TH low: Up/Down/A/Start
	v := b.Read8(0xA10003)
	if v&0x01 != 0 {
		t.Fatalf("Up should read active-low (0) when pressed, got bit=%#02x", v&0x01)
	}
	if v&0x20 != 0 {
		t.Fatalf("Start should read active-low (0) when pressed, got bit=%#02x", v&0x20)
	}

	b.Write8(0xA10003, 0x40) // TH high: Up/Down/B/C
	v = b.Read8(0xA10003)
	if v&0x01 != 0 {
		t.Fatalf("Up should still read active-low (0) when pressed, got bit=%#02x", v&0x01)
	}
	if v&0x40 == 0 {
		t.Fatal("TH bit should echo back what was written")
	}
}
