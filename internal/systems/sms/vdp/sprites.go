package vdp

const maxSpritesPerLine = 8
const spriteTerminatorY = 0xD0 // in 192-line mode, this Y value ends the sprite list early

type spriteHit struct {
	x, tile, row int
}

// evaluateSprites walks the Sprite Attribute Table (64 Y bytes, then 64
// X/tile pairs) looking for sprites that intersect line, stopping at
// the Y=0xD0 terminator or 8 found sprites, whichever comes first.
func (v *VDP) evaluateSprites(line int) ([]spriteHit, bool) {
	base := v.spriteTableBase()
	height := v.spriteHeight()
	shiftLeft := v.regs[0]&0x08 != 0

	var hits []spriteHit
	overflow := false

	for i := 0; i < 64; i++ {
		y := int(v.vram[(base+uint16(i))&0x3FFF])
		if y == spriteTerminatorY {
			break
		}
		spriteY := y + 1
		row := line - spriteY
		if row < 0 || row >= height {
			continue
		}
		if len(hits) >= maxSpritesPerLine {
			overflow = true
			break
		}

		xAddr := base + 0x80 + uint16(i)*2
		x := int(v.vram[xAddr&0x3FFF])
		if shiftLeft {
			x -= 8
		}
		tile := int(v.vram[(xAddr+1)&0x3FFF])
		hits = append(hits, spriteHit{x: x, tile: tile, row: row})
	}
	return hits, overflow
}

// renderSpritesLine composites every sprite found on this scanline: SMS
// sprites have no per-sprite priority or flip bits (unlike NES/GBC) -
// the only priority rule is the background tile's own priority bit, and
// collisions are flagged whenever two *opaque* sprite pixels overlap,
// regardless of which one visually wins.
func (v *VDP) renderSpritesLine(line int, bg *[Width]bgPixel) {
	hits, overflow := v.evaluateSprites(line)
	if overflow {
		v.status |= statusSpriteOverflow
	}

	height := v.spriteHeight()
	patternBase := v.spritePatternBase()
	var opaque [Width]bool

	for _, hit := range hits {
		tile, row := hit.tile, hit.row
		if height == 16 {
			tile &^= 1
			if row >= 8 {
				tile++
				row -= 8
			}
		}
		tileAddr := patternBase + uint16(tile)*32

		for col := 0; col < 8; col++ {
			px := hit.x + col
			if px < 0 || px >= Width {
				continue
			}
			colorIdx := v.decodeTilePixel(tileAddr, col, row)
			if colorIdx == 0 {
				continue
			}
			if opaque[px] {
				v.status |= statusSpriteCollide
				continue // first (lowest SAT index) sprite drawn already owns this pixel
			}
			opaque[px] = true

			if bg[px].priority && bg[px].colorIdx != 0 {
				continue
			}
			r, g, b := v.cram.rgb(16 + colorIdx)
			v.setPixel(px, line, r, g, b)
		}
	}
}
