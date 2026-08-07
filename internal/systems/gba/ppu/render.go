package ppu

func (p *PPU) resolveBG(index uint16) (r, g, b byte) {
	c := p.palette[index&0x1FF]
	return resolveRGB555(c)
}

// renderScanline dispatches on the display mode this project
// implements (0 or 3); any other mode renders as a blank (backdrop)
// screen rather than guessing at unsupported layouts.
func (p *PPU) renderScanline(line int) {
	switch p.mode() {
	case 0:
		p.renderMode0Line(line)
	case 3:
		p.renderMode3Line(line)
	default:
		for x := 0; x < Width; x++ {
			r, g, b := p.resolveBG(0)
			p.frame.SetPixel(x, line, r, g, b, 0xFF)
		}
	}
}

func (p *PPU) renderMode0Line(line int) {
	var spriteIdx [Width]uint16
	var spriteOpaque [Width]bool
	if p.objEnabled() {
		p.spritesLine(line, &spriteIdx, &spriteOpaque)
	}

	for x := 0; x < Width; x++ {
		finalIdx := uint16(0)
		found := false

		// Backgrounds composited back-to-front, BG3 (lowest priority
		// number wins in real hardware via BGxCNT's priority field;
		// this project simplifies to a fixed BG0-over-BG3 stacking
		// order instead of reading each BG's priority field).
		for bg := 3; bg >= 0; bg-- {
			if !p.bgEnabled(bg) {
				continue
			}
			if idx, opaque := p.backgroundPixel(bg, x, line); opaque {
				finalIdx = idx
				found = true
			}
		}

		if spriteOpaque[x] {
			finalIdx = spriteIdx[x]
			found = true
		}

		if !found {
			finalIdx = 0
		}
		r, g, b := p.resolveBG(finalIdx)
		p.frame.SetPixel(x, line, r, g, b, 0xFF)
	}
}

func (p *PPU) renderMode3Line(line int) {
	var spriteIdx [Width]uint16
	var spriteOpaque [Width]bool
	if p.objEnabled() {
		p.spritesLine(line, &spriteIdx, &spriteOpaque)
	}

	for x := 0; x < Width; x++ {
		if spriteOpaque[x] {
			r, g, b := p.resolveBG(spriteIdx[x])
			p.frame.SetPixel(x, line, r, g, b, 0xFF)
			continue
		}
		r, g, b := p.mode3Pixel(x, line)
		p.frame.SetPixel(x, line, r, g, b, 0xFF)
	}
}
