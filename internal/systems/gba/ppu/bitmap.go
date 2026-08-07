package ppu

// mode3Pixel reads Mode 3's single full-screen 15-bit direct-color
// bitmap directly - no palette indirection, unlike every other mode.
func (p *PPU) mode3Pixel(x, y int) (r, g, b byte) {
	addr := uint32(y*Width+x) * 2
	word := p.vramWord(addr)
	return resolveRGB555(word)
}

func resolveRGB555(word uint16) (r, g, b byte) {
	r = expand5to8(byte(word) & 0x1F)
	g = expand5to8(byte(word>>5) & 0x1F)
	b = expand5to8(byte(word>>10) & 0x1F)
	return
}

func expand5to8(v byte) byte { return v<<3 | v>>2 }
