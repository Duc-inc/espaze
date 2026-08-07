package memory

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/input"
)

type fakePPU struct {
	vramAddrLo, vramAddrHi byte
	bgControl              [4]byte
	mainScreen             byte
}

func (f *fakePPU) WriteVRAMAddrLow(v byte)            { f.vramAddrLo = v }
func (f *fakePPU) WriteVRAMAddrHigh(v byte)           { f.vramAddrHi = v }
func (f *fakePPU) WriteVRAMDataLow(v byte)            {}
func (f *fakePPU) WriteVRAMDataHigh(v byte)           {}
func (f *fakePPU) ReadVRAMLow() byte                  { return 0x11 }
func (f *fakePPU) ReadVRAMHigh() byte                 { return 0x22 }
func (f *fakePPU) WriteCGRAMAddr(v byte)              {}
func (f *fakePPU) WriteCGRAMData(v byte)              {}
func (f *fakePPU) WriteOAMByte(addr uint16, v byte)   {}
func (f *fakePPU) ReadOAMByte(addr uint16) byte       { return 0 }
func (f *fakePPU) WriteBGControl(layer int, v byte)   { f.bgControl[layer] = v }
func (f *fakePPU) WriteBGScrollH(layer int, v uint16) {}
func (f *fakePPU) WriteBGScrollV(layer int, v uint16) {}
func (f *fakePPU) WriteMainScreen(v byte)             { f.mainScreen = v }

type fakePorts struct{ data [4]byte }

func (f *fakePorts) Read(i int) byte     { return f.data[i&0x03] }
func (f *fakePorts) Write(i int, v byte) { f.data[i&0x03] = v }

func newTestBus() (*Bus, *fakePPU, *fakePorts) {
	video := &fakePPU{}
	ports := &fakePorts{}
	rom := make([]byte, 0x8000)
	rom[0] = 0xAB
	return New(rom, video, ports), video, ports
}

func TestCartridgeReadThroughBus(t *testing.T) {
	b, _, _ := newTestBus()
	if v := b.Read8(0x008000); v != 0xAB {
		t.Fatalf("Read8(0x8000) = %#02x, want 0xAB", v)
	}
}

func TestWorkRAMMirrorReadWriteRoundTrip(t *testing.T) {
	b, _, _ := newTestBus()
	b.Write8(0x000010, 0x42)
	if v := b.Read8(0x7E0010); v != 0x42 {
		t.Fatalf("Read8(0x7E0010) = %#02x, want 0x42 (should alias the low WRAM mirror)", v)
	}
}

func TestVRAMAddressWriteDelegates(t *testing.T) {
	b, video, _ := newTestBus()
	b.Write8(0x002116, 0x34)
	b.Write8(0x002117, 0x12)
	if video.vramAddrLo != 0x34 || video.vramAddrHi != 0x12 {
		t.Fatalf("VRAM addr = %#02x,%#02x, want 0x34,0x12", video.vramAddrLo, video.vramAddrHi)
	}
}

func TestBGControlAndMainScreenDelegate(t *testing.T) {
	b, video, _ := newTestBus()
	b.Write8(0x002107, 0x01)
	b.Write8(0x00212C, 0x1F)
	if video.bgControl[0] != 0x01 || video.mainScreen != 0x1F {
		t.Fatalf("bgControl[0]=%#02x mainScreen=%#02x, want 0x01/0x1F", video.bgControl[0], video.mainScreen)
	}
}

func TestAudioPortReadWriteDelegates(t *testing.T) {
	b, _, ports := newTestBus()
	b.Write8(0x002140, 0x77)
	if ports.data[0] != 0x77 {
		t.Fatalf("port 0 = %#02x, want 0x77", ports.data[0])
	}
	if v := b.Read8(0x002140); v != 0x77 {
		t.Fatalf("Read8(port 0) = %#02x, want 0x77", v)
	}
}

func TestControllerReadReflectsHeldButtons(t *testing.T) {
	b, _, _ := newTestBus()
	b.SetButtons(input.State{}.With(A, true))
	if v := b.Read8(0x004219); v&(1<<(A-8)) == 0 {
		t.Fatal("expected A button bit set in the high report byte")
	}
}
