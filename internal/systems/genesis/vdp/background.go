package vdp

// renderPlaneLine renders one scanline of scroll plane A or B (planeIsA
// selects which VSRAM/h-scroll-table slot to read). Only whole-screen
// scroll granularity is implemented - real hardware can also scroll
// per 8-line band, per scanline, or (vertically) per column, which a
// handful of games use for parallax and raster effects this won't
// reproduce.
func (v *VDP) renderPlaneLine(line int, base uint16, planeIsA byte, idxOut *[Width]byte, priOut *[Width]bool) {
	width, height := v.planeSize()
	hTableBase := v.hScrollTableBase()

	var hscroll, vscroll uint16
	if planeIsA == 1 {
		hscroll = v.vramWord(hTableBase) & 0x3FF
		vscroll = v.vsram[0] & 0x3FF
	} else {
		hscroll = v.vramWord(hTableBase+2) & 0x3FF
		vscroll = v.vsram[1] & 0x3FF
	}

	planeW, planeH := width*8, height*8
	srcY := (line + int(vscroll)) % planeH
	tileRow := srcY / 8
	withinY := srcY % 8

	for x := 0; x < Width; x++ {
		srcX := ((x-int(hscroll))%planeW + planeW) % planeW
		tileCol := srcX / 8
		withinX := srcX % 8

		entry := v.vramWord(base + uint16(tileRow*width+tileCol)*2)
		tileIdx := entry & 0x07FF
		hFlip := entry&0x0800 != 0
		vFlip := entry&0x1000 != 0
		paletteLine := byte(entry>>13) & 0x03
		priority := entry&0x8000 != 0

		px, py := withinX, withinY
		if hFlip {
			px = 7 - px
		}
		if vFlip {
			py = 7 - py
		}

		idxOut[x] = paletteLine<<4 | v.tilePixel(tileIdx, px, py)
		priOut[x] = priority
	}
}

// tilePixel reads one pixel's 4-bit color index out of an 8x8 tile -
// Genesis tiles are packed two pixels per byte (unlike the NES/SMS
// cores in this project, which use bit-planed tiles), 4 bytes per row.
func (v *VDP) tilePixel(tileIdx uint16, x, y int) byte {
	addr := tileIdx*32 + uint16(y)*4 + uint16(x/2)
	b := v.vram[addr&0xFFFF]
	if x%2 == 0 {
		return b >> 4
	}
	return b & 0x0F
}
