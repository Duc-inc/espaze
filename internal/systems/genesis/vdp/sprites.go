package vdp

const maxSpritesPerFrame = 80
const maxSpritesPerLine = 20

// spriteEntry is one Sprite Attribute Table record, decoded from its 4
// packed words.
type spriteEntry struct {
	y                       int
	widthTiles, heightTiles int
	link                    byte
	priority                bool
	paletteLine             byte
	vFlip, hFlip            bool
	tileIdx                 uint16
	x                       int
}

func (v *VDP) readSprite(index byte) spriteEntry {
	base := v.spriteTableBase() + uint16(index)*8
	w0 := v.vramWord(base)
	w1 := v.vramWord(base + 2)
	w2 := v.vramWord(base + 4)
	w3 := v.vramWord(base + 6)

	return spriteEntry{
		y:           int(w0&0x03FF) - 128,
		heightTiles: int((w1>>8)&0x03) + 1,
		widthTiles:  int((w1>>10)&0x03) + 1,
		link:        byte(w1),
		priority:    w2&0x8000 != 0,
		paletteLine: byte(w2>>13) & 0x03,
		vFlip:       w2&0x1000 != 0,
		hFlip:       w2&0x0800 != 0,
		tileIdx:     w2 & 0x07FF,
		x:           int(w3&0x03FF) - 128,
	}
}

// renderSpritesLine walks the sprite linked list (starting at index 0,
// each entry naming the next; a link of 0 ends the list) looking for
// sprites that intersect line, then composites the ones it finds.
func (v *VDP) renderSpritesLine(line int, idxOut *[Width]byte, priOut, opaqueOut *[Width]bool) {
	index := byte(0)
	found := 0

	for i := 0; i < maxSpritesPerFrame; i++ {
		s := v.readSprite(index)
		h := s.heightTiles * 8
		if line >= s.y && line < s.y+h && found < maxSpritesPerLine {
			v.drawSprite(s, line, idxOut, priOut, opaqueOut)
			found++
		}
		if s.link == 0 {
			break
		}
		index = s.link
	}
}

func (v *VDP) drawSprite(s spriteEntry, line int, idxOut *[Width]byte, priOut, opaqueOut *[Width]bool) {
	row := line - s.y
	if s.vFlip {
		row = s.heightTiles*8 - 1 - row
	}
	tileRow := row / 8
	withinY := row % 8

	for col := 0; col < s.widthTiles*8; col++ {
		px := s.x + col
		if px < 0 || px >= Width {
			continue
		}

		srcCol := col
		if s.hFlip {
			srcCol = s.widthTiles*8 - 1 - col
		}
		tileCol := srcCol / 8
		withinX := srcCol % 8

		// Sprites lay their tiles out column-major (Genesis hardware
		// convention: tile index increases going down a column first,
		// then across to the next column).
		tile := s.tileIdx + uint16(tileCol*s.heightTiles+tileRow)
		colorIdx := v.tilePixel(tile, withinX, withinY)
		if colorIdx == 0 {
			continue
		}

		idxOut[px] = s.paletteLine<<4 | colorIdx
		priOut[px] = s.priority
		opaqueOut[px] = true
	}
}
