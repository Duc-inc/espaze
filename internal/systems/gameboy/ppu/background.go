package ppu

// renderBackgroundLine paints the scrolled background for one scanline
// and records each pixel's raw color index in bgIdx (sprites need it to
// resolve BG-priority).
func (p *PPU) renderBackgroundLine(line int, bgIdx *[Width]byte) {
	tileMapBase := uint16(0x9800)
	if p.lcdc&lcdcBGMap != 0 {
		tileMapBase = 0x9C00
	}
	unsigned := p.lcdc&lcdcTileData != 0

	y := (line + int(p.scy)) & 0xFF
	tileRow := y / 8
	within := y % 8

	for x := 0; x < Width; x++ {
		scrolledX := (x + int(p.scx)) & 0xFF
		tileCol := scrolledX / 8

		mapAddr := tileMapBase + uint16(tileRow)*32 + uint16(tileCol)
		tileIdx := p.vram[mapAddr-0x8000]
		tileAddr := tileDataAddr(tileIdx, unsigned)

		colorIdx := p.tilePixel(tileAddr, scrolledX%8, within)
		bgIdx[x] = colorIdx
		p.setPixel(x, line, applyPalette(colorIdx, p.bgp))
	}
}

// renderWindowLine overlays the window layer where it's visible on this
// scanline, overwriting whatever the background just drew there.
func (p *PPU) renderWindowLine(line int, bgIdx *[Width]byte) {
	windowY := line - int(p.wy)
	if windowY < 0 {
		return
	}

	tileMapBase := uint16(0x9800)
	if p.lcdc&lcdcWindowMap != 0 {
		tileMapBase = 0x9C00
	}
	unsigned := p.lcdc&lcdcTileData != 0

	tileRow := windowY / 8
	within := windowY % 8
	startX := int(p.wx) - 7

	for x := 0; x < Width; x++ {
		if x < startX {
			continue
		}
		winX := x - startX
		tileCol := winX / 8

		mapAddr := tileMapBase + uint16(tileRow)*32 + uint16(tileCol)
		tileIdx := p.vram[mapAddr-0x8000]
		tileAddr := tileDataAddr(tileIdx, unsigned)

		colorIdx := p.tilePixel(tileAddr, winX%8, within)
		bgIdx[x] = colorIdx
		p.setPixel(x, line, applyPalette(colorIdx, p.bgp))
	}
}
