package ppu

// backgroundLine holds one scanline's worth of resolved background
// pixels. rawColor is the tile's own 2-bit color index (0 meaning
// "transparent", i.e. the universal background color) before palette
// lookup - sprite priority and sprite-0 hit both need that raw value
// separately from the final on-screen color.
type backgroundLine struct {
	master   [256]byte
	rawColor [256]byte
}

// renderBackgroundLine renders one visible scanline of background
// pixels from the current scroll position - a batch equivalent of the
// tile-by-tile fetch pipeline real hardware runs across the scanline.
func (p *PPU) renderBackgroundLine() backgroundLine {
	var line backgroundLine

	if p.mask&maskShowBG == 0 {
		universal := p.palette.read(0)
		for x := range line.master {
			line.master[x] = universal
		}
		return line
	}

	coarseXBase := p.scroll.coarseX()
	coarseYBase := p.scroll.coarseY()
	ntXBase := p.scroll.nametableX()
	ntYBase := p.scroll.nametableY()
	fineY := p.scroll.fineY()
	fineX := p.scroll.fineX

	patternBase := uint16(0x0000)
	if p.ctrl&ctrlBGPattern != 0 {
		patternBase = 0x1000
	}

	for x := 0; x < 256; x++ {
		combined := int(fineX) + x
		tileCol := int(coarseXBase) + combined/8
		pixelInTile := combined % 8
		ntXBit := ntXBase
		if tileCol >= 32 {
			tileCol -= 32
			ntXBit ^= 1
		}
		nametableIndex := int(ntYBase)<<1 | int(ntXBit)
		bank := nametableBank(p.mirroring(), nametableIndex)

		ntAddr := uint16(tileCol) + uint16(coarseYBase)*32
		tileIdx := p.nametables[bank][ntAddr]

		attrAddr := uint16(0x3C0) + uint16(coarseYBase/4)*8 + uint16(tileCol/4)
		attrByte := p.nametables[bank][attrAddr]
		quadrant := (coarseYBase%4)/2*2 + byte(tileCol%4)/2
		paletteSel := (attrByte >> (quadrant * 2)) & 0x03

		tileAddr := patternBase + uint16(tileIdx)*16 + uint16(fineY)
		lo := p.cart.ReadCHR(tileAddr)
		hi := p.cart.ReadCHR(tileAddr + 8)
		bit := 7 - pixelInTile
		colorIdx := ((hi>>bit)&1)<<1 | ((lo >> bit) & 1)

		if x < 8 && p.mask&maskShowBGLeft8 == 0 {
			colorIdx = 0
		}

		line.rawColor[x] = colorIdx
		if colorIdx == 0 {
			line.master[x] = p.palette.read(0)
		} else {
			line.master[x] = p.palette.read(uint16(paletteSel)*4 + uint16(colorIdx))
		}
	}
	return line
}
