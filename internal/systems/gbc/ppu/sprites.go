package ppu

const maxSpritesPerLine = 10

// renderSpritesLine paints every sprite visible on this scanline.
// Unlike DMG, CGB priority among overlapping sprites is purely by OAM
// index (lower wins) - X position doesn't factor in, a real hardware
// behavior difference from DMG mode.
func (p *PPU) renderSpritesLine(line int, bg *[Width]bgPixel) {
	if p.lcdc&lcdcOBJEnable == 0 {
		return
	}

	height := 8
	if p.lcdc&lcdcOBJSize != 0 {
		height = 16
	}

	var visible []int
	for i := 0; i < 40 && len(visible) < maxSpritesPerLine; i++ {
		spriteY := int(p.oam[i*4]) - 16
		if line < spriteY || line >= spriteY+height {
			continue
		}
		visible = append(visible, i)
	}

	// Draw lowest-priority (highest OAM index) first so index 0 ends up
	// on top once the loop reaches it.
	for k := len(visible) - 1; k >= 0; k-- {
		p.drawSprite(visible[k], line, height, bg)
	}
}

func (p *PPU) drawSprite(oamIndex, line, height int, bg *[Width]bgPixel) {
	base := oamIndex * 4
	spriteY := int(p.oam[base]) - 16
	spriteX := int(p.oam[base+1]) - 8
	tileIdx := p.oam[base+2]
	attr := p.oam[base+3]

	if height == 16 {
		tileIdx &^= 0x01
	}

	row := line - spriteY
	if attr&0x40 != 0 { // Y flip
		row = height - 1 - row
	}

	tileAddr := uint16(0x8000) + uint16(tileIdx)*16
	if row >= 8 {
		tileAddr += 16
		row -= 8
	}

	bank := (attr >> 3) & 1
	palette := attr & 0x07
	behindBG := attr&0x80 != 0
	xFlip := attr&0x20 != 0
	masterPriority := p.lcdc&lcdcMasterPriority != 0

	for col := 0; col < 8; col++ {
		px := spriteX + col
		if px < 0 || px >= Width {
			continue
		}
		srcCol := col
		if xFlip {
			srcCol = 7 - col
		}
		colorIdx := p.tilePixel(bank, tileAddr, srcCol, row)
		if colorIdx == 0 {
			continue // transparent
		}

		if masterPriority {
			bgWins := bg[px].colorIdx != 0 && (bg[px].priority || behindBG)
			if bgWins {
				continue
			}
		}

		r, g, b := p.objPalettes.color(palette, colorIdx)
		p.setPixel(px, line, r, g, b)
	}
}
