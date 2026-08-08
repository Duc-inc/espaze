package texture

// Texture holds decoded texel colors ready for sampling.
type Texture struct {
	Width, Height int
	texels        []Color
}

// New decodes raw texture data into a sampleable Texture.
func New(f Format, raw []byte, width, height int) *Texture {
	return &Texture{Width: width, Height: height, texels: Decode(f, raw, width, height)}
}

// Sample looks up the nearest texel for a pixel-space coordinate,
// wrapping past the texture's edges - this project's vertex format
// (see gpu.Vertex) carries U/V as raw integers rather than real
// hardware's normalized/fixed-point texture coordinates, so Sample
// takes pixel-space coordinates directly rather than a normalized
// 0-1 range. No filtering: real hardware also supports bilinear,
// this project only does point sampling.
func (t *Texture) Sample(u, v int) Color {
	if t.Width == 0 || t.Height == 0 {
		return Color{}
	}
	x := wrapCoord(u, t.Width)
	y := wrapCoord(v, t.Height)
	return t.texels[y*t.Width+x]
}

func wrapCoord(i, size int) int {
	i %= size
	if i < 0 {
		i += size
	}
	return i
}
