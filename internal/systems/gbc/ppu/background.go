package ppu

// bgPixel is one scanline pixel's resolved background/window state:
// the raw 2-bit color index (0 = transparent for sprite-priority
// purposes) and whether its tile attribute claims priority over sprites.
type bgPixel struct {
	colorIdx byte
	priority bool
}

// renderBackgroundLine paints the scrolled background for one scanline,
// resolving each tile's attribute byte (palette, bank, flips, priority)
// from VRAM bank 1 at the same address as its tile index in bank 0.
func (p *PPU) renderBackgroundLine(line int, bg *[Width]bgPixel) {
	tileMapBase := uint16(0x9800)
	if p.lcdc&lcdcBGMap != 0 {
		tileMapBase = 0x9C00
	}
	unsigned := p.lcdc&lcdcTileData != 0

	y := (line + int(p.scy)) & 0xFF
	tileRow := y / 8

	for x := 0; x < Width; x++ {
		scrolledX := (x + int(p.scx)) & 0xFF
		tileCol := scrolledX / 8

		mapAddr := tileMapBase + uint16(tileRow)*32 + uint16(tileCol)
		tileIdx := p.vram[0][mapAddr-0x8000]
		attr := p.vram[1][mapAddr-0x8000]

		within := y % 8
		col := scrolledX % 8
		if attr&0x40 != 0 { // vertical flip
			within = 7 - within
		}
		if attr&0x20 != 0 { // horizontal flip
			col = 7 - col
		}

		bank := (attr >> 3) & 1
		tileAddr := tileDataAddr(tileIdx, unsigned)
		colorIdx := p.tilePixel(bank, tileAddr, col, within)

		bg[x] = bgPixel{colorIdx: colorIdx, priority: attr&0x80 != 0}
		r, g, b := p.bgPalettes.color(attr&0x07, colorIdx)
		p.setPixel(x, line, r, g, b)
	}
}

// renderWindowLine overlays the window layer where it's visible on this
// scanline, exactly mirroring renderBackgroundLine's attribute handling.
func (p *PPU) renderWindowLine(line int, bg *[Width]bgPixel) {
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
	startX := int(p.wx) - 7

	for x := 0; x < Width; x++ {
		if x < startX {
			continue
		}
		winX := x - startX
		tileCol := winX / 8

		mapAddr := tileMapBase + uint16(tileRow)*32 + uint16(tileCol)
		tileIdx := p.vram[0][mapAddr-0x8000]
		attr := p.vram[1][mapAddr-0x8000]

		within := windowY % 8
		col := winX % 8
		if attr&0x40 != 0 {
			within = 7 - within
		}
		if attr&0x20 != 0 {
			col = 7 - col
		}

		bank := (attr >> 3) & 1
		tileAddr := tileDataAddr(tileIdx, unsigned)
		colorIdx := p.tilePixel(bank, tileAddr, col, within)

		bg[x] = bgPixel{colorIdx: colorIdx, priority: attr&0x80 != 0}
		r, g, b := p.bgPalettes.color(attr&0x07, colorIdx)
		p.setPixel(x, line, r, g, b)
	}
}
