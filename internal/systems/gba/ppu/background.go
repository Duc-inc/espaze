package ppu

// bgCharBase/bgScreenBase/bg8bpp/bgSizeTiles decode a BGxCNT register's
// layout fields.
func bgCharBase(cnt uint16) uint32   { return uint32((cnt>>2)&0x03) * 0x4000 }
func bgScreenBase(cnt uint16) uint32 { return uint32((cnt>>8)&0x1F) * 0x800 }
func bg8bpp(cnt uint16) bool         { return cnt&0x80 != 0 }
func bgSizeTiles(cnt uint16) (w, h int) {
	switch (cnt >> 14) & 0x03 {
	case 1:
		return 64, 32
	case 2:
		return 32, 64
	case 3:
		return 64, 64
	default:
		return 32, 32
	}
}

// bgTilePixel4/8bpp read one pixel's palette-index bits out of a tile,
// applying h/v flip.
func (p *PPU) bgTilePixel(charBase uint32, tileIndex uint16, x, y int, flipH, flipV, use8bpp bool) byte {
	if flipH {
		x = 7 - x
	}
	if flipV {
		y = 7 - y
	}
	if use8bpp {
		addr := charBase + uint32(tileIndex)*64 + uint32(y)*8 + uint32(x)
		return p.vram[addr&0x17FFF]
	}
	addr := charBase + uint32(tileIndex)*32 + uint32(y)*4 + uint32(x/2)
	b := p.vram[addr&0x17FFF]
	if x&1 == 0 {
		return b & 0x0F
	}
	return b >> 4
}

// backgroundPixel resolves one background's color index (0 =
// transparent) at screen (x,y) after scrolling, or false if
// transparent.
func (p *PPU) backgroundPixel(bg int, x, y int) (uint16, bool) {
	cnt := p.bgcnt[bg]
	charBase := bgCharBase(cnt)
	screenBase := bgScreenBase(cnt)
	use8bpp := bg8bpp(cnt)
	tilesW, tilesH := bgSizeTiles(cnt)

	scrollX := (x + int(p.bghofs[bg])) % (tilesW * 8)
	scrollY := (y + int(p.bgvofs[bg])) % (tilesH * 8)

	tileX, tileY := scrollX/8, scrollY/8
	mapW := tilesW
	if mapW > 32 {
		mapW = 32
	}
	// Screen-block layout for sizes beyond 32x32 tiles: each 32x32
	// block is its own 2KB screen block, laid out left-to-right then
	// top-to-bottom.
	blockX, blockY := tileX/32, tileY/32
	blocksPerRow := tilesW / 32
	block := blockY*blocksPerRow + blockX
	entryAddr := screenBase + uint32(block)*0x800 + uint32((tileY%32)*32+(tileX%32))*2

	entry := p.vramWord(entryAddr)
	tileIndex := entry & 0x03FF
	flipH := entry&0x0400 != 0
	flipV := entry&0x0800 != 0
	paletteBank := byte(entry>>12) & 0x0F

	colorBits := p.bgTilePixel(charBase, tileIndex, scrollX%8, scrollY%8, flipH, flipV, use8bpp)
	if colorBits == 0 {
		return 0, false
	}
	if use8bpp {
		return uint16(colorBits), true
	}
	return uint16(paletteBank)<<4 | uint16(colorBits), true
}
