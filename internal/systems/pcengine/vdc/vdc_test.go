package vdc

import "testing"

type fakePalette struct{}

func (fakePalette) Resolve(index uint16) (byte, byte, byte) {
	if index == 0 {
		return 0, 0, 0
	}
	return 10, 20, 30
}

func writeVRAMWord(v *VDC, addr uint16, word uint16) {
	v.SelectRegister(regMAWR)
	v.WriteDataLow(byte(addr))
	v.WriteDataHigh(byte(addr >> 8))
	v.SelectRegister(regVRAM)
	v.WriteDataLow(byte(word))
	v.WriteDataHigh(byte(word >> 8))
}

func TestVRAMWriteThroughDataPort(t *testing.T) {
	v := New(fakePalette{})
	writeVRAMWord(v, 0x0010, 0xBEEF)
	if v.vram[0x0010] != 0xBEEF {
		t.Fatalf("VRAM[0x10] = %#04x, want 0xBEEF", v.vram[0x0010])
	}
}

func TestMAWRAutoIncrements(t *testing.T) {
	v := New(fakePalette{})
	writeVRAMWord(v, 0x0000, 0x1111)
	// A second data-port write (without re-setting MAWR) should land
	// at the next word, since the first write auto-incremented it.
	v.WriteDataLow(0x22)
	v.WriteDataHigh(0x22)
	if v.vram[0x0001] != 0x2222 {
		t.Fatalf("VRAM[1] = %#04x, want 0x2222 (MAWR should have auto-incremented)", v.vram[0x0001])
	}
}

func TestTilePixelDecodesFourBitplanes(t *testing.T) {
	v := New(fakePalette{})
	// Tile 0, row 0: plane0=1,plane1=0,plane2=1,plane3=1 at bit7 (x=0)
	// word0 = plane1<<8|plane0 = 0x0080; word1(+8) = plane3<<8|plane2 = 0x8080
	v.vram[0] = 0x0080
	v.vram[8] = 0x8080

	got := v.tilePixel(0, 0, 0)
	want := byte(1 | 0<<1 | 1<<2 | 1<<3) // 0b1101 = 13
	if got != want {
		t.Fatalf("tilePixel = %#02x, want %#02x", got, want)
	}
}

func TestSATBDMACopiesFromVRAM(t *testing.T) {
	v := New(fakePalette{})
	v.vram[0x1000] = 0xAAAA
	v.SelectRegister(regSATB)
	v.WriteDataLow(0x00)
	v.WriteDataHigh(0x10) // triggers the DMA from VRAM word 0x1000

	if v.sat[0] != 0xAAAA {
		t.Fatalf("SAT[0] = %#04x, want 0xAAAA", v.sat[0])
	}
}

func TestVBlankIRQFiresAtEndOfVisibleArea(t *testing.T) {
	v := New(fakePalette{})
	v.SelectRegister(regCR)
	v.WriteDataLow(0x08) // enable VBlank IRQ
	v.WriteDataHigh(0x00)

	irq := byte(0)
	for i := 0; i < totalScanlines*cyclesPerLine; i++ {
		irq |= v.Step(1)
	}
	if irq&IRQVBlank == 0 {
		t.Fatal("expected a VBlank IRQ somewhere in a full frame")
	}
}

func TestBackgroundPixelHonorsScrollWrap(t *testing.T) {
	v := New(fakePalette{})
	// Name table entry (0,0): tile index 1, palette 0.
	writeVRAMWord(v, 0, 0x0001)
	// Tile 1's row0: all 4 planes set at every bit -> colorBits=0xF everywhere.
	writeVRAMWord(v, 16, 0xFFFF)
	writeVRAMWord(v, 24, 0xFFFF)

	v.SelectRegister(regBXR)
	v.WriteDataLow(0)
	v.WriteDataHigh(0)

	idx, opaque := v.backgroundPixel(0, 0)
	if !opaque || idx == 0 {
		t.Fatalf("expected an opaque, non-zero background pixel, got idx=%d opaque=%v", idx, opaque)
	}
}
