package vdc

// renderScanline resolves one line: sprites marked high-priority draw
// over the background, everything else falls back to the background
// (or the backdrop, VCE palette index 0, where neither is opaque).
func (v *VDC) renderScanline(line int) {
	var spriteIdx [Width]uint16
	var spriteOpaque, spritePri [Width]bool
	if v.spritesEnabled() {
		v.spritesLine(line, &spriteIdx, &spriteOpaque, &spritePri)
	}

	for x := 0; x < Width; x++ {
		var bgIdx uint16
		var bgOpaque bool
		if v.bgEnabled() {
			bgIdx, bgOpaque = v.backgroundPixel(x, line)
		}

		var finalIdx uint16
		switch {
		case spriteOpaque[x] && (spritePri[x] || !bgOpaque):
			finalIdx = spriteIdx[x]
		case bgOpaque:
			finalIdx = bgIdx
		default:
			finalIdx = 0
		}

		r, g, b := v.palette.Resolve(finalIdx)
		v.frame.SetPixel(x, line, r, g, b, 0xFF)
	}
}
