package memory

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/input"
)

type fakePPU struct {
	vram    [0x4000]byte
	control byte
}

func (f *fakePPU) WriteControl(v byte)                 { f.control = v }
func (f *fakePPU) WriteScrollX(v byte)                 {}
func (f *fakePPU) WriteScrollY(v byte)                 {}
func (f *fakePPU) ReadVRAM(addr uint32) byte           { return f.vram[addr&0x3FFF] }
func (f *fakePPU) WriteVRAM(addr uint32, v byte)       { f.vram[addr&0x3FFF] = v }
func (f *fakePPU) ReadSprite(addr uint32) byte         { return 0 }
func (f *fakePPU) WriteSprite(addr uint32, v byte)     {}
func (f *fakePPU) WritePaletteLow(index byte, v byte)  {}
func (f *fakePPU) WritePaletteHigh(index byte, v byte) {}

type fakeZ80 struct{ ram [0x100]byte }

func (f *fakeZ80) Read(addr uint16) byte     { return f.ram[addr&0xFF] }
func (f *fakeZ80) Write(addr uint16, v byte) { f.ram[addr&0xFF] = v }

func newTestBus() (*Bus, *fakePPU, *fakeZ80) {
	video := &fakePPU{}
	z80 := &fakeZ80{}
	rom := make([]byte, 0x1000)
	rom[0] = 0xAB
	return New(rom, video, z80), video, z80
}

func TestCartridgeReadThroughBus(t *testing.T) {
	b, _, _ := newTestBus()
	if v := b.Read8(0); v != 0xAB {
		t.Fatalf("Read8(0) = %#02x, want 0xAB", v)
	}
}

func TestWorkRAMReadWriteRoundTrip(t *testing.T) {
	b, _, _ := newTestBus()
	b.Write16(wramBase, 0x1234)
	if v := b.Read16(wramBase); v != 0x1234 {
		t.Fatalf("Read16 = %#04x, want 0x1234", v)
	}
}

func TestVRAMWriteDelegatesToPPU(t *testing.T) {
	b, video, _ := newTestBus()
	b.Write8(vramBase+5, 0x42)
	if video.vram[5] != 0x42 {
		t.Fatalf("PPU VRAM[5] = %#02x, want 0x42", video.vram[5])
	}
}

func TestZ80WindowDelegatesToBridge(t *testing.T) {
	b, _, z80 := newTestBus()
	b.Write8(z80Base+0x10, 0x77)
	if z80.ram[0x10] != 0x77 {
		t.Fatalf("Z80 RAM[0x10] = %#02x, want 0x77", z80.ram[0x10])
	}
}

func TestControllerReadsActiveLow(t *testing.T) {
	b, _, _ := newTestBus()
	b.SetButtons(input.State{}.With(A, true))
	v := b.Read8(ioBase)
	if v&(1<<A) != 0 {
		t.Fatal("expected A button bit to read active-low (0) when held")
	}
}

func TestZ80ResetLineToggles(t *testing.T) {
	b, _, _ := newTestBus()
	b.Write8(ioBase+4, 0x00) // reset asserted (active low)
	if !b.Z80ResetAsserted() {
		t.Fatal("expected Z80 reset to be asserted")
	}
	b.Write8(ioBase+4, 0x01)
	if b.Z80ResetAsserted() {
		t.Fatal("expected Z80 reset to be released")
	}
}
