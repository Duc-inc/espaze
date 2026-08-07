package tms9918

// spriteEntry is one of the 32 sprite records (Y, X, pattern index,
// color). Every sprite here is a fixed 8x8, single-color - real
// hardware's 16x16/magnified sizes aren't implemented.
type spriteEntry struct {
	y, x  int
	tile  byte
	color byte
}

func (t *TMS9918) readSprite(index int) spriteEntry {
	base := spriteAttrBase + uint16(index*4)
	y := int(t.vram[base&0x3FFF])
	x := int(t.vram[(base+1)&0x3FFF])
	tile := t.vram[(base+2)&0x3FFF]
	color := t.vram[(base+3)&0x3FFF] & 0x0F
	return spriteEntry{y: y + 1, x: x, tile: tile, color: color} // Y is stored as one less than the actual position on real hardware
}

func (t *TMS9918) spritePixel(s spriteEntry, sx, sy int) bool {
	addr := spritePatternBase + uint16(s.tile)*8 + uint16(sy)
	b := t.vram[addr&0x3FFF]
	return b&(1<<uint(7-sx)) != 0
}

// spritesLine renders every sprite touching this scanline, index 0
// drawn last (highest priority among overlaps) - real hardware's
// per-line sprite count limit (and the resulting 5th-sprite flag)
// isn't reproduced.
func (t *TMS9918) spritesLine(line int, colorOut *[Width]byte, opaqueOut *[Width]bool) {
	if !t.spritesEnabled() {
		return
	}
	for i := 31; i >= 0; i-- {
		s := t.readSprite(i)
		if s.color == 0 || line < s.y || line >= s.y+8 {
			continue
		}
		for sx := 0; sx < 8; sx++ {
			screenX := s.x + sx
			if screenX < 0 || screenX >= Width {
				continue
			}
			if !t.spritePixel(s, sx, line-s.y) {
				continue
			}
			colorOut[screenX] = s.color
			opaqueOut[screenX] = true
		}
	}
}
