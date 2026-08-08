package gpu

// TEV (Texture Environment) is Flipper's fixed-function "shader"
// stage: up to 16 configurable stages combining a texture sample and
// vertex/constant colors per pixel. This project implements the
// combiner math for the 4 most common operations, but nothing feeds
// it a real texture sample yet - there's no texture memory, UV
// interpolation, or filtering here, only the arithmetic a real stage
// would apply once those exist. DrawTriangle (render.go) doesn't call
// this yet; it stays purely Gouraud-shaded from vertex color.
type TEVOp byte

const (
	TEVReplace  TEVOp = iota // output = texture color
	TEVModulate              // output = texture color * vertex color
	TEVAdd                   // output = texture color + vertex color
	TEVDecal                 // output = texture color, alpha-blended over vertex color
)

// Color is a simple 8-bit-per-channel RGBA color, the unit TEV
// combines.
type Color struct{ R, G, B, A byte }

// Combine applies one TEV stage's operation to a texture sample and
// the incoming (vertex/previous-stage) color.
func Combine(op TEVOp, texel, incoming Color) Color {
	switch op {
	case TEVReplace:
		return texel
	case TEVModulate:
		return Color{
			R: mulChannel(texel.R, incoming.R),
			G: mulChannel(texel.G, incoming.G),
			B: mulChannel(texel.B, incoming.B),
			A: mulChannel(texel.A, incoming.A),
		}
	case TEVAdd:
		return Color{
			R: addChannel(texel.R, incoming.R),
			G: addChannel(texel.G, incoming.G),
			B: addChannel(texel.B, incoming.B),
			A: addChannel(texel.A, incoming.A),
		}
	default: // TEVDecal
		alpha := float64(texel.A) / 255.0
		return Color{
			R: blendChannel(texel.R, incoming.R, alpha),
			G: blendChannel(texel.G, incoming.G, alpha),
			B: blendChannel(texel.B, incoming.B, alpha),
			A: incoming.A,
		}
	}
}

func mulChannel(a, b byte) byte { return byte(uint16(a) * uint16(b) / 255) }

func addChannel(a, b byte) byte {
	sum := uint16(a) + uint16(b)
	if sum > 255 {
		return 255
	}
	return byte(sum)
}

func blendChannel(texel, incoming byte, alpha float64) byte {
	return byte(float64(texel)*alpha + float64(incoming)*(1-alpha))
}
