// Package vce implements the PC Engine's HuC6260 Video Color Encoder:
// a 512-entry palette (9-bit RGB, 3 bits per channel) the VDC's tile
// and sprite color indices are resolved through to produce the final
// picture. Real hardware also handles the actual NTSC/RGB signal
// output and a global "dot clock" select this project has no use for
// (frames are handed to the frontend as RGB already).
package vce

// VCE holds the 512-color palette memory and the two address/data
// ports software writes it through.
type VCE struct {
	palette [512]uint16 // 9-bit RGB, ---BBBGGGRRR
	addr    uint16
}

// New returns a VCE with every palette entry black.
func New() *VCE { return &VCE{} }

// Reset clears the palette.
func (v *VCE) Reset() { *v = VCE{} }

// WriteAddressLow/WriteAddressHigh implement the palette address port
// (2 bytes covering the 9-bit index into the 512-entry table).
func (v *VCE) WriteAddressLow(b byte)  { v.addr = v.addr&0x100 | uint16(b) }
func (v *VCE) WriteAddressHigh(b byte) { v.addr = v.addr&0x0FF | uint16(b&0x01)<<8 }

// WriteDataLow/WriteDataHigh implement the palette data port; a write
// to the high byte latches the full color and auto-increments the
// address, matching real hardware's two-byte-per-color protocol.
func (v *VCE) WriteDataLow(b byte) {
	v.palette[v.addr&0x1FF] = v.palette[v.addr&0x1FF]&0x100 | uint16(b)
}

func (v *VCE) WriteDataHigh(b byte) {
	idx := v.addr & 0x1FF
	v.palette[idx] = v.palette[idx]&0x0FF | uint16(b&0x01)<<8
	v.addr++
}

// Resolve converts a 9-bit color index into 8-bit-per-channel RGB.
func (v *VCE) Resolve(index uint16) (r, g, b byte) {
	c := v.palette[index&0x1FF]
	r = expand3to8(byte(c) & 0x07)
	g = expand3to8(byte(c>>3) & 0x07)
	b = expand3to8(byte(c>>6) & 0x07)
	return
}

func expand3to8(v byte) byte { return v<<5 | v<<2 | v>>1 }
