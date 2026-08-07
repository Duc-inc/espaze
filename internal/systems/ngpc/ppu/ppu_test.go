package ppu

import "testing"

func TestBackgroundTileDecodeAndPalette(t *testing.T) {
	p := New()
	p.WriteControl(0x01) // BG enabled

	// Name table entry (0,0): tile 1, palette bank 1.
	entryAddr := uint32(0)
	p.WriteVRAM(entryAddr, 1)
	p.WriteVRAM(entryAddr+1, 0x10)

	// Tile 1's row 0: pixel 0 = color index 7 (low nibble of first byte).
	tileAddr := uint32(1) * bytesPerTile
	p.WriteVRAM(tileAddr, 0x07)

	idx, opaque := p.backgroundPixel(0, 0)
	if !opaque {
		t.Fatal("expected an opaque pixel")
	}
	want := uint16(1)<<4 | 7
	if idx != want {
		t.Fatalf("palette index = %#04x, want %#04x", idx, want)
	}
}

func TestPaletteWriteAndResolve(t *testing.T) {
	p := New()
	p.WritePaletteLow(0, 0x0F) // blue nibble
	p.WritePaletteHigh(0, 0x00)
	_, _, b := p.resolveColor(0)
	if b == 0 {
		t.Fatal("expected a non-zero blue channel")
	}
}

func TestSpriteFlipHReversesPixelOrder(t *testing.T) {
	p := New()
	// Sprite tile 0: row 0, pixel 0 = color 3, pixel 7 = color 0.
	p.WriteVRAM(spriteVRAMBase, 0x03)

	s := spriteEntry{tile: 0, flipH: false}
	if v := p.spritePixel(s, 0, 0); v != 3 {
		t.Fatalf("unflipped pixel 0 = %d, want 3", v)
	}
	s.flipH = true
	if v := p.spritePixel(s, 7, 0); v != 3 {
		t.Fatalf("flipped pixel 7 = %d, want 3", v)
	}
}

func TestVBlankIRQFiresAfterVisibleLines(t *testing.T) {
	p := New()
	irq := byte(0)
	for i := 0; i < totalScanlines*cyclesPerLine; i++ {
		irq |= p.Step(1)
	}
	if irq&IRQVBlank == 0 {
		t.Fatal("expected a VBlank IRQ somewhere in a full frame")
	}
}
