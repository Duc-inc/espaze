package ppu

// spriteVRAMBase keeps sprite tile patterns out of the background
// layers' own tile range.
const spriteVRAMBase = 0x4000

// spriteEntry is one of the 128 sprite records, read from a flat
// 4-bytes-per-sprite layout (X, Y, tile, attribute) - real hardware's
// separate high table (extending X to 9 bits and selecting one of two
// sprite sizes) isn't modeled, so every sprite here is 0-255 in X and
// a fixed 8x8, on top of this package's other documented simplifications.
type spriteEntry struct {
	x, y         int
	tile         byte
	palette      byte
	flipH, flipV bool
}

func (p *PPU) readSprite(index int) spriteEntry {
	base := uint16(index * 4)
	x := int(p.oam[base])
	y := int(p.oam[base+1])
	tile := p.oam[base+2]
	attr := p.oam[base+3]
	return spriteEntry{
		x: x, y: y, tile: tile,
		palette: attr & 0x07,
		flipH:   attr&0x40 != 0,
		flipV:   attr&0x80 != 0,
	}
}

func (p *PPU) spritePixel(s spriteEntry, sx, sy int) byte {
	if s.flipH {
		sx = 7 - sx
	}
	if s.flipV {
		sy = 7 - sy
	}
	base := spriteVRAMBase + uint32(s.tile)*16
	row01 := p.vram[(base+uint32(sy))&0xFFFF]
	row23 := p.vram[(base+uint32(sy)+8)&0xFFFF]
	bit := uint(7 - sx)
	p0 := byte(row01) >> bit & 1
	p1 := byte(row01>>8) >> bit & 1
	p2 := byte(row23) >> bit & 1
	p3 := byte(row23>>8) >> bit & 1
	return p0 | p1<<1 | p2<<2 | p3<<3
}

// spritesLine renders every sprite touching this scanline, OAM index
// 0 drawn last (highest priority among overlaps).
func (p *PPU) spritesLine(line int, idxOut *[Width]uint16, opaqueOut *[Width]bool) {
	for i := 127; i >= 0; i-- {
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
			idxOut[screenX] = 128 + uint16(s.palette)*16 + uint16(colorBits)
			opaqueOut[screenX] = true
		}
	}
}
