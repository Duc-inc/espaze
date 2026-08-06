package ppu

// cgbPalettes implements the two 64-byte palette RAMs (background and
// object, each 8 palettes of 4 RGB555 colors) reachable through an
// auto-incrementing index/data port pair - BCPS/BCPD for background,
// OCPS/OCPD for objects, both wired to the same logic here.
type cgbPalettes struct {
	data     [64]byte
	index    byte
	autoIncr bool
}

func (p *cgbPalettes) writeIndex(v byte) {
	p.index = v & 0x3F
	p.autoIncr = v&0x80 != 0
}

func (p *cgbPalettes) readIndexPort() byte {
	v := p.index
	if p.autoIncr {
		v |= 0x80
	}
	return v
}

func (p *cgbPalettes) readData() byte { return p.data[p.index] }

func (p *cgbPalettes) writeData(v byte) {
	p.data[p.index] = v
	if p.autoIncr {
		p.index = (p.index + 1) & 0x3F
	}
}

// color resolves one of a palette's 4 colors to RGB888. Each color is
// two bytes, little-endian RGB555 (bits 0-4 red, 5-9 green, 10-14
// blue); bit-replicating the top bits into the low ones is the standard
// 5-to-8-bit expansion used to avoid colors looking dimmer than they
// should.
func (p *cgbPalettes) color(paletteIdx, colorIdx byte) (r, g, b byte) {
	offset := int(paletteIdx)*8 + int(colorIdx)*2
	lo := p.data[offset]
	hi := p.data[offset+1]
	raw := uint16(hi)<<8 | uint16(lo)

	r5 := byte(raw & 0x1F)
	g5 := byte((raw >> 5) & 0x1F)
	b5 := byte((raw >> 10) & 0x1F)

	return expand5to8(r5), expand5to8(g5), expand5to8(b5)
}

func expand5to8(v byte) byte { return v<<3 | v>>2 }
