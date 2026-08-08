// registers.go holds the XF unit's control registers: which
// position/normal matrix index a vertex uses, how many texture-
// coordinate generators are active, and the projection matrix and
// viewport scale/offset - state real hardware keeps separate from XF
// memory proper (memory.go).
package xf

// ProjectionType selects between the two projection shapes the GX API
// (and so real XF projection registers) support.
type ProjectionType uint32

const (
	ProjectionPerspective  ProjectionType = 0
	ProjectionOrthographic ProjectionType = 1
)

// Projection holds the XF unit's projection state the way real
// hardware stores it: not a raw 4x4 matrix, but a small set of
// coefficients plus a type selecting how they combine.
type Projection struct {
	Coeffs [6]float32
	Type   ProjectionType
}

// Matrix builds the Mat4 this project's transform pipeline
// (transform.go) actually multiplies vertices against, from
// Projection's raw coefficients. This is this project's own
// best-effort mapping from a 6-coefficient perspective/orthographic
// projection to a standard 4x4 - derived from general projection-
// matrix construction, not yet cross-checked field-by-field against a
// specific public register table, so treat the exact coefficient
// ordering as provisional until verified.
func (p Projection) Matrix() Mat4 {
	c := p.Coeffs
	if p.Type == ProjectionOrthographic {
		return Mat4{
			{c[0], 0, 0, c[1]},
			{0, c[2], 0, c[3]},
			{0, 0, c[4], c[5]},
			{0, 0, 0, 1},
		}
	}
	return Mat4{
		{c[0], 0, c[1], 0},
		{0, c[2], c[3], 0},
		{0, 0, c[4], c[5]},
		{0, 0, -1, 0},
	}
}

// Viewport maps normalized device coordinates (post perspective
// divide) onto screen pixel coordinates, mirroring real hardware's
// own scale+offset viewport registers.
type Viewport struct {
	ScaleX, ScaleY   float32
	OffsetX, OffsetY float32
}

// Registers holds the XF unit's per-vertex control state: which
// matrices to use, how many texture coordinate sets to generate, the
// projection, and the viewport mapping to screen coordinates.
type Registers struct {
	PosMatrixIndex    uint16 // base address into Memory for the active PosMatrix
	NormalMatrixIndex uint16 // base address into Memory for the active NormalMatrix
	TexGenCount       int    // active texture coordinate generators (0-8 on real hardware)

	Projection Projection
	Viewport   Viewport
}

// NewRegisters returns Registers in a default state: no texgens
// active, and a viewport that leaves normalized coordinates
// unchanged.
func NewRegisters() Registers {
	return Registers{
		Viewport: Viewport{ScaleX: 1, ScaleY: 1},
	}
}
