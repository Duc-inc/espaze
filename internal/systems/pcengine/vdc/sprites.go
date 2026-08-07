package vdc

// spriteEntry is one of the SAT's 64 4-word records: Y position, X
// position, pattern index, and an attribute word (palette, priority,
// flip). Every sprite this project renders is a fixed 16x16 pixels -
// real hardware's 16x32/32x16/32x64 size options aren't implemented.
type spriteEntry struct {
	y, x      int
	tileIndex uint16
	palette   byte
	priority  bool
	flipH     bool
	flipV     bool
}

func (v *VDC) readSprite(index int) spriteEntry {
	base := index * 4
	yWord := v.sat[base]
	xWord := v.sat[base+1]
	patternWord := v.sat[base+2]
	attr := v.sat[base+3]

	return spriteEntry{
		y:         int(yWord&0x3FF) - 64, // real hardware offsets sprite Y by 64 lines
		x:         int(xWord&0x3FF) - 32, // and X by 32 columns, both to allow off-screen positioning
		tileIndex: (patternWord >> 1) & 0x3FF,
		palette:   byte(attr) & 0x0F,
		priority:  attr&0x80 != 0,
		flipH:     attr&0x0800 != 0,
		flipV:     attr&0x2000 != 0,
	}
}

// spritePixel returns a 16x16 sprite's color bits at its-local (sx,sy)
// (0-15 each), reading its 4 8x8 sub-tiles - stored, by this
// project's best understanding of the real layout, column-major: tile
// 0 top-left, tile 1 bottom-left, tile 2 top-right, tile 3 bottom-right.
func (v *VDC) spritePixel(s spriteEntry, sx, sy int) byte {
	if s.flipH {
		sx = 15 - sx
	}
	if s.flipV {
		sy = 15 - sy
	}

	col := sx / 8
	row := sy / 8
	subTile := s.tileIndex*4 + uint16(col*2+row)
	return v.tilePixel(subTile, sx%8, sy%8)
}

// spritesLine renders every enabled sprite touching this scanline into
// idxOut/priOut, in SAT order (index 0 drawn last, i.e. highest
// priority among overlapping sprites) - real hardware's actual
// per-sprite ordering/overflow behavior isn't fully reproduced.
func (v *VDC) spritesLine(line int, idxOut *[Width]uint16, opaqueOut, priOut *[Width]bool) {
	for i := 63; i >= 0; i-- {
		s := v.readSprite(i)
		if line < s.y || line >= s.y+16 {
			continue
		}
		for sx := 0; sx < 16; sx++ {
			screenX := s.x + sx
			if screenX < 0 || screenX >= Width {
				continue
			}
			colorBits := v.spritePixel(s, sx, line-s.y)
			if colorBits == 0 {
				continue
			}
			idxOut[screenX] = 256 + uint16(s.palette)<<4 | uint16(colorBits)
			opaqueOut[screenX] = true
			priOut[screenX] = s.priority
		}
	}
}
