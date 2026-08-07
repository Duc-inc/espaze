package ppu

func (p *PPU) resolveColor(index uint16) (r, g, b byte) {
	c := p.palette[index&0x1F]
	r = expand4to8(byte(c>>8) & 0x0F)
	g = expand4to8(byte(c>>4) & 0x0F)
	b = expand4to8(byte(c) & 0x0F)
	return
}

func expand4to8(v byte) byte { return v<<4 | v }

// renderScanline composites the background and sprites - sprites
// always draw above the background, a simplification of the real
// hardware's per-sprite/per-tile priority bits.
func (p *PPU) renderScanline(line int) {
	var spriteIdx [Width]uint16
	var spriteOpaque [Width]bool
	if p.objEnable {
		p.spritesLine(line, &spriteIdx, &spriteOpaque)
	}

	for x := 0; x < Width; x++ {
		var idx uint16
		if spriteOpaque[x] {
			idx = spriteIdx[x]
		} else if p.bgEnable {
			if bgIdx, opaque := p.backgroundPixel(x, line); opaque {
				idx = bgIdx
			}
		}
		r, g, b := p.resolveColor(idx)
		p.frame.SetPixel(x, line, r, g, b, 0xFF)
	}
}
