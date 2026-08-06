package ppu

// evaluatedSprite is one sprite the scanline evaluator picked out of
// OAM, with row already resolved to "which row of its pattern data" -
// accounting for the sprite's own Y position and vertical flip.
type evaluatedSprite struct {
	tile, attr, x byte
	oamIndex      int
	row           int
}

// evaluateSprites finds up to 8 sprites that intersect scanline,
// reporting overflow if more than 8 do. Sprite Y in OAM is the scanline
// *before* the sprite starts (hardware takes one line to latch it),
// hence the +1 here.
func evaluateSprites(o *oam, scanline, spriteHeight int) ([]evaluatedSprite, bool) {
	var found []evaluatedSprite
	overflow := false

	for i := 0; i < 64; i++ {
		base := i * 4
		y := o.data[base]
		row := scanline - (int(y) + 1)
		if row < 0 || row >= spriteHeight {
			continue
		}

		attr := o.data[base+2]
		if attr&0x80 != 0 { // vertical flip
			row = spriteHeight - 1 - row
		}

		if len(found) >= 8 {
			overflow = true
			break
		}
		found = append(found, evaluatedSprite{
			tile: o.data[base+1], attr: attr, x: o.data[base+3],
			oamIndex: i, row: row,
		})
	}
	return found, overflow
}

// spritePixel is one x-coordinate's resolved sprite pixel for a scanline.
type spritePixel struct {
	master    byte
	opaque    bool
	behindBG  bool
	isSprite0 bool
}

// renderSpriteLine composites every sprite found on this scanline into
// per-pixel output, respecting OAM priority (lower index wins).
func (p *PPU) renderSpriteLine(sprites []evaluatedSprite) [256]spritePixel {
	var line [256]spritePixel
	if p.mask&maskShowSprites == 0 {
		return line
	}

	height := 8
	if p.ctrl&ctrlSpriteSize8x16 != 0 {
		height = 16
	}
	patternBase := uint16(0x0000)
	if height == 8 && p.ctrl&ctrlSpritePattern != 0 {
		patternBase = 0x1000
	}

	// Draw back-to-front so the lowest OAM index (highest priority) ends
	// up as the topmost pixel once the loop reaches it.
	for i := len(sprites) - 1; i >= 0; i-- {
		s := sprites[i]
		row := s.row
		tile := s.tile
		base := patternBase
		if height == 16 {
			base = uint16(tile&1) * 0x1000
			tile &^= 1
			if row >= 8 {
				tile++
				row -= 8
			}
		}

		lo := p.cart.ReadCHR(base + uint16(tile)*16 + uint16(row))
		hi := p.cart.ReadCHR(base + uint16(tile)*16 + uint16(row) + 8)

		for col := 0; col < 8; col++ {
			screenX := int(s.x) + col
			if screenX > 255 {
				continue
			}
			if screenX < 8 && p.mask&maskShowSpritesLeft8 == 0 {
				continue
			}

			bit := 7 - col
			if s.attr&0x40 != 0 { // horizontal flip
				bit = col
			}
			colorIdx := ((hi>>bit)&1)<<1 | ((lo >> bit) & 1)
			if colorIdx == 0 {
				continue
			}

			paletteSel := s.attr & 0x03
			line[screenX] = spritePixel{
				master:    p.palette.read(0x10 + uint16(paletteSel)*4 + uint16(colorIdx)),
				opaque:    true,
				behindBG:  s.attr&0x20 != 0,
				isSprite0: s.oamIndex == 0,
			}
		}
	}
	return line
}
