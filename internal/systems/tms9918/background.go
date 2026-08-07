package tms9918

const nameTableTiles = 32 // 32x24 tiles = 256x192 pixels, no scrolling

// backgroundPixel resolves the background color index at screen
// (x,y): each name-table position (not each 8-pixel group, unlike
// real Graphics Mode I) gets its own foreground/background color
// pair from the color table - a simplification closer to Graphics
// Mode II's finer granularity than Mode I's coarse one.
func (t *TMS9918) backgroundPixel(x, y int) byte {
	tileX, tileY := x/8, y/8
	tileIndex := t.vram[(nameTableBase+uint16(tileY*nameTableTiles+tileX))&0x3FFF]

	patternAddr := patternTableBase + uint16(tileIndex)*8 + uint16(y%8)
	patternByte := t.vram[patternAddr&0x3FFF]
	bit := uint(7 - x%8)
	set := patternByte&(1<<bit) != 0

	colorAddr := colorTableBase + uint16(tileY*nameTableTiles+tileX)
	colorByte := t.vram[colorAddr&0x3FFF]
	fg := colorByte >> 4
	bg := colorByte & 0x0F

	if set {
		return fg
	}
	return bg
}
