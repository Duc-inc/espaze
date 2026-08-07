package vdp

// cram is the VDP's 32-byte color RAM: two 16-color palettes
// (background, then sprite), each entry a 6-bit color (--BBGGRR - 2
// bits per channel, the SMS VDP's native format).
type cram struct {
	data [32]byte
}

func (c *cram) read(addr byte) byte     { return c.data[addr&0x1F] }
func (c *cram) write(addr byte, v byte) { c.data[addr&0x1F] = v & 0x3F }

// rgb expands one CRAM entry's 2-bit-per-channel color into RGB888 by
// bit-replicating each channel across the byte (00->0x00, 01->0x55,
// 10->0xAA, 11->0xFF), the standard 2-to-8-bit expansion.
func (c *cram) rgb(addr byte) (r, g, b byte) {
	v := c.read(addr)
	r = expand2to8(v & 0x03)
	g = expand2to8((v >> 2) & 0x03)
	b = expand2to8((v >> 4) & 0x03)
	return
}

func expand2to8(v byte) byte { return v * 0x55 }
