package tia

import "math"

// palette holds the 128 NTSC colors the TIA's 7-bit COLUxx registers
// (4-bit hue, 3-bit luminance) select from, computed rather than
// hand-transcribed from a lookup table: hue 0 is a neutral gray ramp
// (luminance only), hues 1-15 are spaced evenly around the NTSC color
// wheel and decoded the same way a composite TV would from a
// chroma/luma signal. This approximates real hardware's color
// generator circuit rather than reproducing its exact analog
// characteristics.
var palette [128][3]byte

func init() {
	for hue := 0; hue < 16; hue++ {
		for lum := 0; lum < 8; lum++ {
			palette[hue<<3|lum] = ntscColor(hue, lum)
		}
	}
}

func ntscColor(hue, lum int) [3]byte {
	brightness := 40.0 + float64(lum)*27.0 // 0-7 -> roughly 40-229

	if hue == 0 {
		v := clampByte(brightness)
		return [3]byte{v, v, v}
	}

	// Hues 1-15 spread around the color wheel, phase-offset so hue 1
	// lands on a gold/orange (the TIA's traditional "hue 1") rather than
	// starting at red.
	angle := (float64(hue-1)/15.0)*2*math.Pi - 0.7
	chroma := 55.0

	r := brightness + chroma*math.Cos(angle)
	g := brightness + chroma*math.Cos(angle-2*math.Pi/3)
	b := brightness + chroma*math.Cos(angle+2*math.Pi/3)
	return [3]byte{clampByte(r), clampByte(g), clampByte(b)}
}

func clampByte(v float64) byte {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return byte(v)
}

// rgb looks up a 7-bit COLUxx value's color (bit0 is unused on real
// hardware).
func rgb(colu byte) (r, g, b byte) {
	c := palette[(colu>>1)&0x7F]
	return c[0], c[1], c[2]
}
