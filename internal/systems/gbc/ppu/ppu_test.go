package ppu

import "testing"

func TestVBKSelectsVRAMBank(t *testing.T) {
	p := New()

	p.WriteRegister(0xFF4F, 0x00)
	p.WriteVRAM(0x8000, 0x11)
	p.WriteRegister(0xFF4F, 0x01)
	p.WriteVRAM(0x8000, 0x22)

	if p.vram[0][0] != 0x11 {
		t.Fatalf("bank 0 = %#02x, want 0x11", p.vram[0][0])
	}
	if p.vram[1][0] != 0x22 {
		t.Fatalf("bank 1 = %#02x, want 0x22", p.vram[1][0])
	}
}

func TestPaletteRGB555Expansion(t *testing.T) {
	var pal cgbPalettes
	// Pure red at max intensity: R=31 (0b11111), G=0, B=0.
	pal.data[0] = 0x1F
	pal.data[1] = 0x00

	r, g, b := pal.color(0, 0)
	if r != 255 {
		t.Fatalf("red = %d, want 255 (max 5-bit value expands to max 8-bit)", r)
	}
	if g != 0 || b != 0 {
		t.Fatalf("g,b = %d,%d, want 0,0", g, b)
	}
}

func TestPaletteAutoIncrement(t *testing.T) {
	var pal cgbPalettes
	pal.writeIndex(0x80) // index 0, auto-increment on

	pal.writeData(0x11)
	pal.writeData(0x22)

	if pal.data[0] != 0x11 || pal.data[1] != 0x22 {
		t.Fatalf("data[0:2] = %#02x %#02x, want 0x11 0x22", pal.data[0], pal.data[1])
	}
	if pal.index != 2 {
		t.Fatalf("index after 2 writes = %d, want 2", pal.index)
	}
}

func TestPaletteIndexPortReportsAutoIncrementFlag(t *testing.T) {
	var pal cgbPalettes
	pal.writeIndex(0x85)
	if got := pal.readIndexPort(); got != 0x85 {
		t.Fatalf("index port = %#02x, want 0x85 (index 5, auto-increment bit set)", got)
	}
}

func TestBackgroundTileAttributeFlipAndPalette(t *testing.T) {
	p := New()
	p.lcdc = lcdcEnable | lcdcTileData // unsigned ($8000-based) tile addressing

	// Tile 1's pattern: row 0 = 0b10000000 in both bit planes -> leftmost
	// pixel is color index 3, everything else 0. Tile 1 starts at $8010;
	// row 0 is its first two bytes (low plane, then high plane).
	p.vram[0][0x10] = 0x80
	p.vram[0][0x11] = 0x80

	// Nametable entry (0,0) lives at $9800, i.e. VRAM offset $1800 - not
	// offset 0, which is inside the tile data region instead.
	const nametableEntry0 = 0x1800
	p.vram[0][nametableEntry0] = 1 // tile index, in bank 0
	// Its attribute (bank 1, same address): horizontal flip + palette 2.
	p.vram[1][nametableEntry0] = 0x22 // bit5 (flip X) | palette bits 0-2 = 2

	p.bgPalettes.data[2*8+3*2] = 0x1F   // palette 2, color 3, low byte: R=31
	p.bgPalettes.data[2*8+3*2+1] = 0x00 // high byte: G=B=0

	var bg [Width]bgPixel
	p.renderBackgroundLine(0, &bg)

	// With horizontal flip, the tile's leftmost source pixel (color 3)
	// should land on screen X=7, not X=0.
	if bg[7].colorIdx != 3 {
		t.Fatalf("bg[7].colorIdx = %d, want 3 (flipped tile)", bg[7].colorIdx)
	}
	if r := p.frame.Pixels[7*4]; r != 255 {
		t.Fatalf("pixel(7,0) red = %d, want 255", r)
	}
}
