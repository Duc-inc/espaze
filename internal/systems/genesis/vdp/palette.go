package vdp

// cram is the VDP's 128-byte color RAM: 4 palettes of 16 colors, each
// color a 16-bit word in the format ----BBB-GGG-RRR- (3 bits per
// channel; the unused bits read back as zero on real hardware, and
// this project doesn't bother modeling the "mode" register's shadow/
// highlight brightness variants some games use for lighting effects).
type cram struct {
	data [64]uint16
}

func (c *cram) write(index byte, v uint16) { c.data[index&0x3F] = v & 0x0EEE }
func (c *cram) read(index byte) uint16     { return c.data[index&0x3F] }

func (c *cram) rgb(index byte) (r, g, b byte) {
	v := c.read(index)
	r = expand3to8(byte(v>>1) & 0x07)
	g = expand3to8(byte(v>>5) & 0x07)
	b = expand3to8(byte(v>>9) & 0x07)
	return
}

func expand3to8(v byte) byte { return v<<5 | v<<2 | v>>1 }
