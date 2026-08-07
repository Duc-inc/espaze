package tms9918

// backdropColor returns register 7's low nibble, the color drawn
// where nothing else is opaque.
func (t *TMS9918) backdropColor() byte { return t.regs[7] & 0x0F }

func (t *TMS9918) renderScanline(line int) {
	if !t.displayEnabled() {
		for x := 0; x < Width; x++ {
			t.frame.SetPixel(x, line, 0, 0, 0, 0xFF)
		}
		return
	}

	var spriteColor [Width]byte
	var spriteOpaque [Width]bool
	t.spritesLine(line, &spriteColor, &spriteOpaque)

	for x := 0; x < Width; x++ {
		idx := t.backgroundPixel(x, line)
		if idx == 0 {
			idx = t.backdropColor()
		}
		if spriteOpaque[x] {
			idx = spriteColor[x]
		}
		r, g, b := colorRGB(idx)
		t.frame.SetPixel(x, line, r, g, b, 0xFF)
	}
}
