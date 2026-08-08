// lighting.go implements the XF unit's per-vertex lighting: an
// ambient term plus each enabled light's Lambertian (N.L, clamped to
// zero for surfaces facing away) diffuse contribution. Real hardware
// supports up to 8 lights (GX's own LIGHT0-LIGHT7) with specular
// terms and several attenuation/spot modes; this project models only
// ambient + per-light diffuse from a point position, the simplest
// case that still produces real shading.
package xf

// MaxLights matches GX's own light count (LIGHT0-LIGHT7).
const MaxLights = 8

// LightColor is a light or material's color, accumulated in floating
// point rather than 8-bit-per-channel - lighting math naturally
// overshoots 1.0 before the final result is clamped back to a
// displayable range.
type LightColor struct {
	R, G, B float32
}

// Add returns the sum of two light colors - how multiple lights'
// contributions accumulate.
func (c LightColor) Add(other LightColor) LightColor {
	return LightColor{R: c.R + other.R, G: c.G + other.G, B: c.B + other.B}
}

// MulScalar scales a light color by a factor (e.g. an N.L attenuation
// term).
func (c LightColor) MulScalar(s float32) LightColor {
	return LightColor{R: c.R * s, G: c.G * s, B: c.B * s}
}

// Mul returns the component-wise product of two colors - how a
// surface's material color filters incoming light.
func (c LightColor) Mul(other LightColor) LightColor {
	return LightColor{R: c.R * other.R, G: c.G * other.G, B: c.B * other.B}
}

// Clamp01 clamps every channel to [0, 1], the range a displayable
// color must end up in after accumulating potentially-overshooting
// light contributions.
func (c LightColor) Clamp01() LightColor {
	return LightColor{R: clamp01(c.R), G: clamp01(c.G), B: clamp01(c.B)}
}

func clamp01(v float32) float32 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}

// Light is a single point light source: a position in the same space
// the transformed normal/vertex live in (camera space, matching
// TransformNormal/TransformPosition's output), a color, and whether
// it's currently enabled. Real hardware also supports directional and
// spot lights with distance attenuation curves - not modeled here.
type Light struct {
	Position Vec3
	Color    LightColor
	Enabled  bool
}

// Ambient is the constant, direction-independent light term applied
// regardless of surface orientation - real hardware's ambient color
// register.
type Ambient struct {
	Color LightColor
}

// Illuminate computes the lit color at a vertex: the ambient term
// plus every enabled light's Lambertian diffuse contribution, all
// filtered by the surface's material color and clamped to a
// displayable range.
func Illuminate(position, normal Vec3, material LightColor, ambient Ambient, lights []Light) LightColor {
	sum := ambient.Color
	for _, l := range lights {
		if !l.Enabled {
			continue
		}
		toLight := l.Position.Sub(position).Normalize()
		ndotl := normal.Dot(toLight)
		if ndotl < 0 {
			ndotl = 0
		}
		sum = sum.Add(l.Color.MulScalar(ndotl))
	}
	return sum.Mul(material).Clamp01()
}
