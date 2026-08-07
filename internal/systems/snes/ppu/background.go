package ppu

// Background layout this project fixes rather than reading from
// separate tilemap/character-base registers: a 32x32-tile name table
// at VRAM word address 0, and each tile's own pattern data at word
// address tileIndex*16 (4bpp) or tileIndex*8 (2bpp) - the same "tile
// index directly addresses pattern data" convention this project's PC
// Engine VDC uses. A 4bpp tile's first 8 words hold bitplanes 0-1
// (one row per word, low/high byte), the next 8 hold bitplanes 2-3;
// a 2bpp tile is just the first 8 words.
const nameTableTiles = 32

func (p *PPU) tilePixel(tileIndex uint16, x, y int, use4bpp bool) byte {
	base := uint32(tileIndex) * 8
	if use4bpp {
		base = uint32(tileIndex) * 16
	}
	row01 := p.vram[(base+uint32(y))&0xFFFF]
	bit := uint(7 - x)
	p0 := byte(row01) >> bit & 1
	p1 := byte(row01>>8) >> bit & 1
	if !use4bpp {
		return p0 | p1<<1
	}
	row23 := p.vram[(base+uint32(y)+8)&0xFFFF]
	p2 := byte(row23) >> bit & 1
	p3 := byte(row23>>8) >> bit & 1
	return p0 | p1<<1 | p2<<2 | p3<<3
}

// backgroundPixel resolves one layer's palette index (0 = transparent)
// at screen (x,y) after scrolling.
func (p *PPU) backgroundPixel(layer int, x, y int) (uint16, bool) {
	use4bpp := p.bg4bpp(layer)
	scrollX := (x + int(p.bgHOfs[layer])) % (nameTableTiles * 8)
	scrollY := (y + int(p.bgVOfs[layer])) % (nameTableTiles * 8)
	tileX, tileY := scrollX/8, scrollY/8

	entryAddr := uint32(tileY*nameTableTiles + tileX)
	entry := p.vram[entryAddr&0xFFFF]
	tileIndex := entry & 0x03FF
	paletteGroup := (entry >> 10) & 0x07
	flipX := entry&0x4000 != 0
	flipY := entry&0x8000 != 0

	px, py := scrollX%8, scrollY%8
	if flipX {
		px = 7 - px
	}
	if flipY {
		py = 7 - py
	}

	colorBits := p.tilePixel(tileIndex, px, py, use4bpp)
	if colorBits == 0 {
		return 0, false
	}
	if use4bpp {
		return paletteGroup*16 + uint16(colorBits), true
	}
	return paletteGroup*4 + uint16(colorBits), true
}
