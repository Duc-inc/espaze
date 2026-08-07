package ppu

import "testing"

func TestMode3PixelReadsDirectColorBitmap(t *testing.T) {
	p := New()
	p.WriteDISPCNT(0x0003) // mode 3
	// Write a 15-bit color word directly into VRAM at (0,0).
	word := uint16(0x1F) // red only
	p.vram[0], p.vram[1] = byte(word), byte(word>>8)

	r, _, _ := p.mode3Pixel(0, 0)
	if r == 0 {
		t.Fatal("expected a non-zero red channel from the bitmap")
	}
}

func TestBackgroundPixelDecodesTileAndPalette(t *testing.T) {
	p := New()
	// BG0: char base 0, screen base 0, 4bpp, 32x32 tiles.
	p.WriteBGCNT(0, 0x0000)

	// Name table entry at (0,0): tile 1, palette bank 2.
	entry := uint16(1) | uint16(2)<<12
	p.vram[0x0000], p.vram[0x0001] = byte(entry), byte(entry>>8)

	// Tile 1's row 0: pixel 0 = color index 5 (low nibble of first byte).
	tileAddr := uint32(1) * 32
	p.vram[tileAddr] = 0x05

	idx, opaque := p.backgroundPixel(0, 0, 0)
	if !opaque {
		t.Fatal("expected an opaque pixel")
	}
	want := uint16(2)<<4 | 5
	if idx != want {
		t.Fatalf("palette index = %#04x, want %#04x", idx, want)
	}
}

func TestSpriteDisabledBitHidesNonAffineSprite(t *testing.T) {
	p := New()
	// attr0: Y=0, disable bit (bit9) set, affine bit (bit8) clear.
	attr0 := uint16(0x0200)
	p.oam[0], p.oam[1] = byte(attr0), byte(attr0>>8)

	s := p.readSprite(0)
	if !s.disabled {
		t.Fatal("expected the sprite to be disabled")
	}
}

func TestVBlankIRQFiresAfterVisibleLines(t *testing.T) {
	p := New()
	p.WriteDISPSTAT(0x08) // enable VBlank IRQ

	irq := byte(0)
	for i := 0; i < totalScanlines*dotsPerLine*cyclesPerDot; i++ {
		irq |= p.Step(1)
	}
	if irq&IRQVBlank == 0 {
		t.Fatal("expected a VBlank IRQ somewhere in a full frame")
	}
}

func TestPaletteWriteAndResolve(t *testing.T) {
	p := New()
	p.WritePalette16(0, 0x001F) // red only, index 0
	r, _, _ := p.resolveBG(0)
	if r == 0 {
		t.Fatal("expected a non-zero red channel after writing palette index 0")
	}
}
