package ppu

// Background layout this project fixes: a 32x32-tile name table at
// VRAM address 0, 2 bytes/entry (12-bit tile index + 4-bit palette
// bank), with each tile's own packed-4bpp 8x8 pattern (2 pixels/byte,
// 32 bytes/tile) stored at tileIndex*32, directly addressed the same
// way this project's PC Engine VDC lays tiles out.
const nameTableTiles = 32
const bytesPerTile = 32

func (p *PPU) tilePixel(tileIndex uint16, x, y int) byte {
	addr := uint32(tileIndex)*bytesPerTile + uint32(y)*4 + uint32(x/2)
	b := p.vram[addr&0x3FFF]
	if x&1 == 0 {
		return b & 0x0F
	}
	return b >> 4
}

// backgroundPixel resolves the background's palette index (0 =
// transparent) at screen (x,y) after scrolling.
func (p *PPU) backgroundPixel(x, y int) (uint16, bool) {
	scrollX := (x + int(p.scrollX)) % (nameTableTiles * 8)
	scrollY := (y + int(p.scrollY)) % (nameTableTiles * 8)
	tileX, tileY := scrollX/8, scrollY/8

	entryAddr := uint32(tileY*nameTableTiles+tileX) * 2
	lo := p.vram[entryAddr&0x3FFF]
	hi := p.vram[(entryAddr+1)&0x3FFF]
	entry := uint16(lo) | uint16(hi)<<8
	tileIndex := entry & 0x0FFF
	paletteBank := byte(entry>>12) & 0x0F

	colorBits := p.tilePixel(tileIndex, scrollX%8, scrollY%8)
	if colorBits == 0 {
		return 0, false
	}
	return uint16(paletteBank)<<4 | uint16(colorBits), true
}
