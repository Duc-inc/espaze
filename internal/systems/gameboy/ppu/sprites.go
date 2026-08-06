package ppu

import "sort"

const maxSpritesPerLine = 10

type spriteHit struct {
	x        int
	oamIndex int
}

// renderSpritesLine paints every sprite visible on this scanline, highest
// priority last so it visually wins: DMG priority is lowest X first, ties
// broken by lowest OAM index (i.e. earliest-found sprite).
func (p *PPU) renderSpritesLine(line int, bgIdx *[Width]byte) {
	height := 8
	if p.lcdc&lcdcOBJSize != 0 {
		height = 16
	}

	var visible []spriteHit
	for i := 0; i < 40 && len(visible) < maxSpritesPerLine; i++ {
		spriteY := int(p.oam[i*4]) - 16
		if line < spriteY || line >= spriteY+height {
			continue
		}
		visible = append(visible, spriteHit{x: int(p.oam[i*4+1]) - 8, oamIndex: i})
	}

	sort.SliceStable(visible, func(a, b int) bool { return visible[a].x < visible[b].x })

	for k := len(visible) - 1; k >= 0; k-- {
		p.drawSprite(visible[k], line, height, bgIdx)
	}
}

func (p *PPU) drawSprite(hit spriteHit, line, height int, bgIdx *[Width]byte) {
	base := hit.oamIndex * 4
	spriteY := int(p.oam[base]) - 16
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

	palette := p.obp0
	if attr&0x10 != 0 {
		palette = p.obp1
	}
	behindBG := attr&0x80 != 0
	xFlip := attr&0x20 != 0

	for col := 0; col < 8; col++ {
		px := hit.x + col
		if px < 0 || px >= Width {
			continue
		}
		srcCol := col
		if xFlip {
			srcCol = 7 - col
		}
		colorIdx := p.tilePixel(tileAddr, srcCol, row)
		if colorIdx == 0 {
			continue // transparent
		}
		if behindBG && bgIdx[px] != 0 {
			continue // background color 1-3 wins
		}
		p.setPixel(px, line, applyPalette(colorIdx, palette))
	}
}
