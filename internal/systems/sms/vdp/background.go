package vdp

// bgPixel is one scanline pixel's resolved background state: the raw
// color index (0 = transparent, showing the backdrop) and whether its
// tile's priority bit claims to draw over sprites.
type bgPixel struct {
	colorIdx byte
	priority bool
}

// renderBackgroundLine paints the scrolled, tile-attribute-aware
// background for one scanline. Each name table entry is 2 bytes: a
// 9-bit tile index, H/V flip, a palette-select bit (choosing between
// CRAM's two 16-color banks), and a priority bit.
func (v *VDP) renderBackgroundLine(line int, bg *[Width]bgPixel) {
	base := v.nameTableBase()
	hScroll := v.hScroll()
	vScroll := v.vScroll()
	lockCol0 := v.regs[0]&reg0HScrollLock != 0
	lockRow := v.regs[0]&reg0VScrollLock != 0

	for x := 0; x < Width; x++ {
		effHScroll := hScroll
		if lockCol0 && x < 8 {
			effHScroll = 0
		}
		srcX := byte(x) - effHScroll

		effVScroll := vScroll
		if lockRow && x >= 192 {
			effVScroll = 0
		}
		srcY := (line + int(effVScroll)) % 224

		tileCol := int(srcX) / 8
		tileRow := srcY / 8
		withinX := int(srcX) % 8
		withinY := srcY % 8

		entryAddr := base + uint16(tileRow*32+tileCol)*2
		lo := v.vram[entryAddr&0x3FFF]
		hi := v.vram[(entryAddr+1)&0x3FFF]

		tileIdx := uint16(hi&0x01)<<8 | uint16(lo)
		if hi&0x02 != 0 { // horizontal flip
			withinX = 7 - withinX
		}
		if hi&0x04 != 0 { // vertical flip
			withinY = 7 - withinY
		}
		paletteSel := (hi >> 3) & 1
		priority := hi&0x10 != 0

		colorIdx := v.decodeTilePixel(tileIdx*32, withinX, withinY)
		bg[x] = bgPixel{colorIdx: colorIdx, priority: priority}

		if colorIdx == 0 {
			r, g, b := v.cram.rgb(v.backdropColor())
			v.setPixel(x, line, r, g, b)
		} else {
			r, g, b := v.cram.rgb(paletteSel*16 + colorIdx)
			v.setPixel(x, line, r, g, b)
		}
	}
}
