package ppu

// spriteVRAMBase is where sprite tile patterns start, keeping them
// out of the background's own tile range.
const spriteVRAMBase = 0x2000

// spriteEntry is one of the 64 sprite table records: Y, X, tile index,
// and an attribute byte (palette bank bit, h/v flip). Every sprite
// this project renders is a fixed 8x8 - the real hardware's larger
// sprite sizes aren't implemented, on top of this package's other
// documented simplifications.
type spriteEntry struct {
	y, x         int
	tile         byte
	flipH, flipV bool
}

func (p *PPU) readSprite(index int) spriteEntry {
	base := uint32(index * 4)
	y := int(p.sprites[base])
	x := int(p.sprites[base+1])
	tile := p.sprites[base+2]
	attr := p.sprites[base+3]
	return spriteEntry{y: y, x: x, tile: tile, flipH: attr&0x01 != 0, flipV: attr&0x02 != 0}
}

func (p *PPU) spritePixel(s spriteEntry, sx, sy int) byte {
	if s.flipH {
		sx = 7 - sx
	}
	if s.flipV {
		sy = 7 - sy
	}
	addr := spriteVRAMBase + uint32(s.tile)*bytesPerTile + uint32(sy)*4 + uint32(sx/2)
	b := p.vram[addr&0x3FFF]
	if sx&1 == 0 {
		return b & 0x0F
	}
	return b >> 4
}

// spritesLine renders every sprite touching this scanline, OAM index 0
// drawn last (highest priority among overlaps).
func (p *PPU) spritesLine(line int, idxOut *[Width]uint16, opaqueOut *[Width]bool) {
	for i := 63; i >= 0; i-- {
		s := p.readSprite(i)
		if line < s.y || line >= s.y+8 {
			continue
		}
		for sx := 0; sx < 8; sx++ {
			screenX := s.x + sx
			if screenX < 0 || screenX >= Width {
				continue
			}
			colorBits := p.spritePixel(s, sx, line-s.y)
			if colorBits == 0 {
				continue
			}
			idxOut[screenX] = 16 + uint16(colorBits)
			opaqueOut[screenX] = true
		}
	}
}
