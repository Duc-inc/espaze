// registers.go holds the XF unit's control-register block. Public
// GameCube hardware notes (YAGCD chapter 5, "internal XF Registers")
// describe this starting at word address 0x1000. Every register's raw
// word is kept (Raw); fields this project has a confirmed bit layout
// for (matrix selection, viewport, projection, texgen.go's TexGen/
// PostTexGen) are decoded too - fields whose bit layout isn't
// confirmed, or that this project doesn't act on yet (lighting-channel
// control), stay raw-only rather than guessed at.
package xf

import "math"

const (
	RegistersStart = 0x1000
	RegistersEnd   = 0x1058 // 0x1057 is the last documented register
)

const registerCount = RegistersEnd - RegistersStart

// XF control register addresses. Descriptions follow public hardware
// notes; several are stored raw-only for now (see package doc).
const (
	RegErrorStatus    = 0x1000
	RegDiagnostics    = 0x1001
	RegInternalState0 = 0x1002
	RegInternalState1 = 0x1003
	RegClockControl   = 0x1004
	RegClipDisable    = 0x1005
	RegPerfSelector0  = 0x1006
	RegPerfSelector1  = 0x1007

	RegVertexSpec       = 0x1008
	RegNumColorOutputs  = 0x1009
	RegAmbientColor0    = 0x100a
	RegAmbientColor1    = 0x100b
	RegMaterialColor0   = 0x100c
	RegMaterialColor1   = 0x100d
	RegColorOutputCtrl0 = 0x100e
	RegColorOutputCtrl1 = 0x100f
	RegAlphaOutputCtrl0 = 0x1010
	RegAlphaOutputCtrl1 = 0x1011
	RegDualTexEnable    = 0x1012
	// 0x1013-0x1017: unknown, raw-only.

	RegMatrixSelection0 = 0x1018
	RegMatrixSelection1 = 0x1019

	RegViewportStart   = 0x101a // 0x101a..0x101f, 6 words
	RegProjectionStart = 0x1020 // 0x1020..0x1025, 6 coefficients
	RegProjectionMode  = 0x1026
	// 0x1027-0x103e: unknown, raw-only.

	RegNumTexGens        = 0x103f
	RegTexCoordCtrlStart = 0x1040 // 0x1040..0x1047, TexGen (texgen.go)
	// 0x1048-0x104f: unknown, raw-only.
	RegPostTexCtrlStart = 0x1050 // 0x1050..0x1057, PostTexGen (texgen.go)
)

// ProjectionType selects between the two projection shapes GX uses.
type ProjectionType uint32

const (
	ProjectionPerspective  ProjectionType = 0
	ProjectionOrthographic ProjectionType = 1
)

// Projection holds XF's six raw projection coefficients plus the
// projection-mode register.
type Projection struct {
	Coeffs [6]float32
	Type   ProjectionType
}

// Matrix builds the 4x4 projection matrix this project's transform
// pipeline (transform.go) multiplies vertices against, from
// Projection's six coefficients.
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

// Viewport mirrors XF's six viewport registers (0x101a..0x101f): an
// X/Y/Z scale and an X/Y/Z offset. Only ScaleX/ScaleY/OffsetX/OffsetY
// are consumed by Apply (transform.go) so far; ScaleZ/OffsetZ are
// decoded and stored but not yet used by any Z-axis mapping.
type Viewport struct {
	ScaleX, ScaleY, ScaleZ    float32
	OffsetX, OffsetY, OffsetZ float32
}

// Registers holds the XF unit's decoded control state.
type Registers struct {
	Raw [registerCount]uint32

	MatrixSelection0 MatrixSelection0
	MatrixSelection1 MatrixSelection1

	// Ambient decodes RegAmbientColor0 as a packed RGBA word, the same
	// convention lightmemory.go's ReadLight uses for light colors.
	// RegAmbientColor1 (the second color channel real hardware also
	// exposes) stays raw-only, matching this project's single-channel
	// lighting model (lighting.go's Illuminate takes one Ambient).
	Ambient Ambient

	Viewport   Viewport
	Projection Projection

	// NumTexGens is RegNumTexGens' decoded value: how many of TexGen's 8
	// slots are actually active. Defaults to 0 (no active generator),
	// matching real hardware's power-on state and keeping vertex UVs as
	// plain stream pass-through until a game configures at least one.
	NumTexGens uint32

	TexGen     [8]TexGen
	PostTexGen [8]PostTexGen
}

// NewRegisters returns a default state that leaves coordinates
// unchanged until a command stream uploads real XF registers.
func NewRegisters() Registers {
	return Registers{
		Viewport: Viewport{ScaleX: 1, ScaleY: 1, ScaleZ: 1},
	}
}

// Write stores one XF control-register word and updates the decoded
// field this project currently models. Unknown or not-yet-decoded
// registers remain available through Raw.
func (r *Registers) Write(addr uint16, word uint32) {
	if addr < RegistersStart || addr >= RegistersEnd {
		return
	}
	r.Raw[addr-RegistersStart] = word

	switch {
	case addr == RegMatrixSelection0:
		r.MatrixSelection0 = MatrixSelection0(word)
	case addr == RegMatrixSelection1:
		r.MatrixSelection1 = MatrixSelection1(word)
	case addr >= RegViewportStart && addr < RegViewportStart+6:
		r.writeViewport(int(addr-RegViewportStart), math.Float32frombits(word))
	case addr >= RegProjectionStart && addr < RegProjectionStart+6:
		r.Projection.Coeffs[addr-RegProjectionStart] = math.Float32frombits(word)
	case addr == RegProjectionMode:
		r.Projection.Type = ProjectionType(word & 1)
	case addr == RegNumTexGens:
		r.NumTexGens = word
	case addr == RegAmbientColor0:
		r.Ambient.Color = rgba8ToLightColor(word)
	case addr >= RegTexCoordCtrlStart && addr < RegTexCoordCtrlStart+8:
		r.TexGen[addr-RegTexCoordCtrlStart] = TexGen{Raw: word}
	case addr >= RegPostTexCtrlStart && addr < RegPostTexCtrlStart+8:
		r.PostTexGen[addr-RegPostTexCtrlStart] = PostTexGen{Raw: word}
	}
}

func (r *Registers) writeViewport(offset int, v float32) {
	switch offset {
	case 0:
		r.Viewport.ScaleX = v
	case 1:
		r.Viewport.ScaleY = v
	case 2:
		r.Viewport.ScaleZ = v
	case 3:
		r.Viewport.OffsetX = v
	case 4:
		r.Viewport.OffsetY = v
	case 5:
		r.Viewport.OffsetZ = v
	}
}
