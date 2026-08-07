package ppu

func (p *PPU) resolveColor(index uint16) (r, g, b byte) {
	c := p.cgram[index&0xFF]
	r = expand5to8(byte(c) & 0x1F)
	g = expand5to8(byte(c>>5) & 0x1F)
	b = expand5to8(byte(c>>10) & 0x1F)
	return
}

func expand5to8(v byte) byte { return v<<3 | v>>2 }

// renderScanline composites background layers 0-3 (layer 0 highest
// priority among them) and sprites, sprites always drawn above every
// background - a simplification of real hardware's per-tile/per-
// sprite priority bit and Main/Sub screen layering.
func (p *PPU) renderScanline(line int) {
	var spriteIdx [Width]uint16
	var spriteOpaque [Width]bool
	if p.objMain {
		p.spritesLine(line, &spriteIdx, &spriteOpaque)
	}

	for x := 0; x < Width; x++ {
		var idx uint16
		for layer := 3; layer >= 0; layer-- {
			if !p.bgEnabled(layer) {
				continue
			}
			if bgIdx, opaque := p.backgroundPixel(layer, x, line); opaque {
				idx = bgIdx
			}
		}
		if spriteOpaque[x] {
			idx = spriteIdx[x]
		}

		r, g, b := p.resolveColor(idx)
		p.frame.SetPixel(x, line, r, g, b, 0xFF)
	}
}
