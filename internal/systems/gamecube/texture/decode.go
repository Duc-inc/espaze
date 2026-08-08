package texture

// Decode converts width*height texels of raw, already-uncompressed
// texture data into a flat, row-major slice of Color. Truncated input
// (fewer bytes than width*height*BytesPerTexel(f)) simply leaves the
// remaining texels zeroed rather than panicking.
func Decode(f Format, raw []byte, width, height int) []Color {
	out := make([]Color, width*height)
	stride := BytesPerTexel(f)
	for i := range out {
		off := i * stride
		if off+stride > len(raw) {
			break
		}
		out[i] = decodeTexel(f, raw[off:])
	}
	return out
}

func decodeTexel(f Format, b []byte) Color {
	switch f {
	case FormatI8:
		return Color{R: b[0], G: b[0], B: b[0], A: 255}
	case FormatRGB565:
		word := uint16(b[0])<<8 | uint16(b[1])
		r5 := byte(word>>11) & 0x1F
		g6 := byte(word>>5) & 0x3F
		b5 := byte(word) & 0x1F
		return Color{
			// Scale each channel up to 8 bits by replicating its high
			// bits into the low bits, the standard N-to-8-bit expansion.
			R: r5<<3 | r5>>2,
			G: g6<<2 | g6>>4,
			B: b5<<3 | b5>>2,
			A: 255,
		}
	default: // FormatRGBA8
		return Color{R: b[0], G: b[1], B: b[2], A: b[3]}
	}
}
