package vdc

// Background tile map layout this project fixes rather than reading
// from MWR: a 32x32-tile (256x256 pixel) name table living at VRAM
// word address 0, wrapping for scroll purposes. Each name table entry
// packs a 12-bit tile index (bits 0-11) and a 4-bit palette select
// (bits 12-15); a tile's own 8x8 4bpp pattern data lives at VRAM word
// address tileIndex*16, the same fixed relationship real hardware
// uses (tile index directly addresses pattern data - there's no
// separate "pattern table base" register to speak of).
const nameTableTiles = 32
const tileWordsPerTile = 16

func (v *VDC) tilePixel(tileIndex uint16, x, y int) byte {
	base := tileIndex * tileWordsPerTile
	row0 := v.vram[(base+uint16(y))&0x7FFF]
	row1 := v.vram[(base+uint16(y)+8)&0x7FFF]

	bit := uint(7 - x)
	p0 := byte(row0) >> bit & 1
	p1 := byte(row0>>8) >> bit & 1
	p2 := byte(row1) >> bit & 1
	p3 := byte(row1>>8) >> bit & 1
	return p0 | p1<<1 | p2<<2 | p3<<3
}

// backgroundPixel returns the palette index (0 = transparent/backdrop)
// at screen position (x,y) after applying the BXR/BYR scroll registers.
func (v *VDC) backgroundPixel(x, y int) (uint16, bool) {
	scrollX := (x + int(v.regs[regBXR])) % (nameTableTiles * 8)
	scrollY := (y + int(v.regs[regBYR])) % (nameTableTiles * 8)
	if scrollX < 0 {
		scrollX += nameTableTiles * 8
	}
	if scrollY < 0 {
		scrollY += nameTableTiles * 8
	}

	tileX, tileY := scrollX/8, scrollY/8
	entry := v.vram[uint16(tileY*nameTableTiles+tileX)&0x7FFF]
	tileIndex := entry & 0x0FFF
	paletteSel := byte(entry>>12) & 0x0F

	colorBits := v.tilePixel(tileIndex, scrollX%8, scrollY%8)
	if colorBits == 0 {
		return 0, false
	}
	return uint16(paletteSel)<<4 | uint16(colorBits), true
}
