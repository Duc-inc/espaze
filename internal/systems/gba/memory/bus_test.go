package memory

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/input"
)

type fakePPU struct {
	vram    [0x18000]byte
	dispcnt uint16
}

func (f *fakePPU) WriteDISPCNT(v uint16)                { f.dispcnt = v }
func (f *fakePPU) ReadDISPCNT() uint16                  { return f.dispcnt }
func (f *fakePPU) WriteDISPSTAT(v uint16)               {}
func (f *fakePPU) WriteBGCNT(bg int, v uint16)          {}
func (f *fakePPU) WriteBGHOFS(bg int, v uint16)         {}
func (f *fakePPU) WriteBGVOFS(bg int, v uint16)         {}
func (f *fakePPU) ReadVRAM8(addr uint32) byte           { return f.vram[addr&0x17FFF] }
func (f *fakePPU) WriteVRAM8(addr uint32, v byte)       { f.vram[addr&0x17FFF] = v }
func (f *fakePPU) ReadOAM8(addr uint32) byte            { return 0 }
func (f *fakePPU) WriteOAM8(addr uint32, v byte)        {}
func (f *fakePPU) ReadPalette8(addr uint32) byte        { return 0 }
func (f *fakePPU) WritePalette16(addr uint32, v uint16) {}

type fakeAPU struct{ fifoA []byte }

func (f *fakeAPU) WriteFIFOA(v byte)       { f.fifoA = append(f.fifoA, v) }
func (f *fakeAPU) WriteFIFOB(v byte)       {}
func (f *fakeAPU) WriteSoundCntH(v uint16) {}

func newTestBus() (*Bus, *fakePPU, *fakeAPU) {
	video := &fakePPU{}
	sound := &fakeAPU{}
	rom := make([]byte, 0x1000)
	rom[0] = 0xAB
	return New(rom, video, sound), video, sound
}

func TestCartridgeReadThroughBus(t *testing.T) {
	b, _, _ := newTestBus()
	if v := b.Read8(0x08000000); v != 0xAB {
		t.Fatalf("Read8 = %#02x, want 0xAB", v)
	}
}

func TestEWRAMReadWriteRoundTrip(t *testing.T) {
	b, _, _ := newTestBus()
	b.Write32(0x02000000, 0xDEADBEEF)
	if v := b.Read32(0x02000000); v != 0xDEADBEEF {
		t.Fatalf("Read32 = %#08x, want 0xDEADBEEF", v)
	}
}

func TestVRAMWriteDelegatesToPPU(t *testing.T) {
	b, video, _ := newTestBus()
	b.Write16(0x06000000, 0x1234)
	if video.vram[0] != 0x34 || video.vram[1] != 0x12 {
		t.Fatalf("PPU VRAM[0:2] = %#02x,%#02x, want 0x34,0x12", video.vram[0], video.vram[1])
	}
}

func TestFIFOAByteWriteDelegatesToAPU(t *testing.T) {
	b, _, sound := newTestBus()
	b.Write8(0x040000A0, 0x42)
	if len(sound.fifoA) != 1 || sound.fifoA[0] != 0x42 {
		t.Fatalf("APU fifoA = %v, want [0x42]", sound.fifoA)
	}
}

func TestImmediateDMATransfersData(t *testing.T) {
	b, _, _ := newTestBus()
	b.Write32(0x02000000, 0xCAFEBABE)

	b.Write32(0x040000B0, 0x02000000) // DMA0 source
	b.Write32(0x040000B4, 0x02001000) // DMA0 dest
	b.Write16(0x040000B8, 1)          // count = 1 word
	b.Write16(0x040000BA, 0x8400)     // enable, 32-bit, immediate start

	if v := b.Read32(0x02001000); v != 0xCAFEBABE {
		t.Fatalf("DMA dest = %#08x, want 0xCAFEBABE", v)
	}
}

func TestTimerOverflowRaisesIRQ(t *testing.T) {
	b, _, _ := newTestBus()
	b.Write16(0x04000100, 0xFFFE) // TM0 reload: overflow after 2 ticks
	b.Write16(0x04000102, 0x00C0) // enable, IRQ on, prescaler /1
	b.Write16(0x04000200, 0x0008) // IE: timer0
	b.Write16(0x04000208, 1)      // IME on

	b.StepTimers(3)
	if !b.InterruptPending() {
		t.Fatal("expected a pending interrupt after timer0 overflows")
	}
}

func TestKeypadReadsActiveLow(t *testing.T) {
	b, _, _ := newTestBus()
	b.SetButtons(input.State{}.With(A, true))
	v := b.Read16(0x04000130)
	if v&0x01 != 0 {
		t.Fatal("expected A button bit to read active-low (0) when held")
	}
	if v&0x02 == 0 {
		t.Fatal("expected B button bit to read released (1)")
	}
}
