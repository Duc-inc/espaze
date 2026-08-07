package ppu

// objVRAMBase is where sprite tile data starts - the upper half of
// VRAM, shared with bitmap-mode background data in modes 3-5 (this
// project doesn't restrict the boundary any further than that).
const objVRAMBase = 0x10000

var spriteDims = [3][4][2]int{
	{{8, 8}, {16, 16}, {32, 32}, {64, 64}}, // square
	{{16, 8}, {32, 8}, {32, 16}, {64, 32}}, // horizontal
	{{8, 16}, {8, 32}, {16, 32}, {32, 64}}, // vertical
}

type spriteEntry struct {
	y, x         int
	w, h         int
	tileIndex    uint16
	paletteBank  byte
	use8bpp      bool
	flipH, flipV bool
	priority     byte
	disabled     bool
}

func (p *PPU) readSprite(index int) spriteEntry {
	base := uint32(index * 8)
	attr0 := p.oamWord(base)
	attr1 := p.oamWord(base + 2)
	attr2 := p.oamWord(base + 4)

	affine := attr0&0x0100 != 0
	disabled := !affine && attr0&0x0200 != 0

	shape := (attr0 >> 14) & 0x03
	size := (attr1 >> 14) & 0x03
	w, h := 8, 8
	if shape < 3 {
		w, h = spriteDims[shape][size][0], spriteDims[shape][size][1]
	}

	y := int(attr0 & 0xFF)
	if y >= 160 {
		y -= 256 // Y is effectively signed for off-top-of-screen positioning
	}
	x := int(attr1 & 0x01FF)
	if x >= 240 {
		x -= 512
	}

	return spriteEntry{
		y: y, x: x, w: w, h: h,
		tileIndex:   attr2 & 0x03FF,
		paletteBank: byte(attr2>>12) & 0x0F,
		use8bpp:     attr0&0x2000 != 0,
		flipH:       !affine && attr1&0x1000 != 0,
		flipV:       !affine && attr1&0x2000 != 0,
		priority:    byte(attr2>>10) & 0x03,
		disabled:    disabled,
	}
}

// spritePixel reads one sprite's color bits at its local (sx,sy),
// using 1D tile mapping only (see this package's doc comment).
func (p *PPU) spritePixel(s spriteEntry, sx, sy int) byte {
	if s.flipH {
		sx = s.w - 1 - sx
	}
	if s.flipV {
		sy = s.h - 1 - sy
	}
	tileCol, tileRow := sx/8, sy/8
	tilesPerRow := s.w / 8

	if s.use8bpp {
		tile := s.tileIndex/2 + uint16(tileRow*tilesPerRow+tileCol)
		addr := objVRAMBase + uint32(tile)*64 + uint32(sy%8)*8 + uint32(sx%8)
		return p.vram[addr&0x17FFF]
	}
	tile := s.tileIndex + uint16(tileRow*tilesPerRow+tileCol)
	addr := objVRAMBase + uint32(tile)*32 + uint32(sy%8)*4 + uint32((sx%8)/2)
	b := p.vram[addr&0x17FFF]
	if sx%2 == 0 {
		return b & 0x0F
	}
	return b >> 4
}

// spritesLine renders every enabled sprite touching this scanline,
// OAM index 0 drawn last (highest priority among overlapping sprites
// of equal priority) - real hardware's exact tie-break and the
// priority field's interaction with backgrounds is simplified to
// "sprites always drawn above backgrounds".
func (p *PPU) spritesLine(line int, idxOut *[Width]uint16, opaqueOut *[Width]bool) {
	for i := 127; i >= 0; i-- {
		s := p.readSprite(i)
		if s.disabled || line < s.y || line >= s.y+s.h {
			continue
		}
		for sx := 0; sx < s.w; sx++ {
			screenX := s.x + sx
			if screenX < 0 || screenX >= Width {
				continue
			}
			colorBits := p.spritePixel(s, sx, line-s.y)
			if colorBits == 0 {
				continue
			}
			if s.use8bpp {
				idxOut[screenX] = 256 + uint16(colorBits)
			} else {
				idxOut[screenX] = 256 + uint16(s.paletteBank)<<4 | uint16(colorBits)
			}
			opaqueOut[screenX] = true
		}
	}
}
