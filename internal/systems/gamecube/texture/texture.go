// Package texture implements Flipper's texture format decoding and
// sampling: converting raw uploaded texture bytes into per-texel
// colors, and looking a color up by pixel coordinate. Real Flipper
// hardware supports many more formats (paletted formats with a
// separate color-lookup table, and an S3TC-like compressed format)
// and stores texels in small tiled blocks rather than simple
// row-major scanline order, with both point and bilinear filtering.
// This project covers three common uncompressed formats (I8
// intensity, RGB565, RGBA8), decodes them row-major rather than
// tile-accurate, and only does nearest-neighbor sampling - documented
// simplifications, not an attempt at cycle-accurate texture hardware.
package texture

// Color is a simple 8-bit-per-channel RGBA color.
type Color struct{ R, G, B, A byte }

// Format selects how raw texel bytes are interpreted.
type Format byte

const (
	FormatI8     Format = iota // 1 byte/texel: intensity, used for R/G/B; alpha always opaque
	FormatRGB565               // 2 bytes/texel, big-endian 5-6-5; alpha always opaque
	// FormatRGBA8 packs 4 bytes/texel as R,G,B,A directly - this
	// project's own simplified linear layout. Real hardware instead
	// splits RGBA8 into separate AR and GB tile planes; that layout
	// isn't modeled here.
	FormatRGBA8
)

// BytesPerTexel returns how many raw bytes one texel occupies in the
// given format.
func BytesPerTexel(f Format) int {
	switch f {
	case FormatI8:
		return 1
	case FormatRGB565:
		return 2
	default: // FormatRGBA8
		return 4
	}
}
