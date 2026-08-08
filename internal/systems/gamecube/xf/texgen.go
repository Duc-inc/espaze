// texgen.go decodes XF's texture-coordinate generator registers
// (RegTexCoordCtrlStart..+8, TEX0-TEX7) and post-transform texture
// registers (RegPostTexCtrlStart..+8, DUALTEX0-7) - both documented
// with an explicit bit table in public hardware notes. TexGen's low 4
// bits (INPUT_FORM/PROJECTION) are the one part of that table this
// project doesn't decode: the source's own bit numbering for those
// two fields wasn't legible enough to implement with confidence, so
// they stay reachable only through Registers.Raw.
package xf

// SourceRow selects which row of incoming vertex data a regular
// (non-embossed) texture-coordinate generator reads from.
type SourceRow uint32

const (
	SourceGeom      SourceRow = 0
	SourceNormal    SourceRow = 1
	SourceColors    SourceRow = 2
	SourceBinormalT SourceRow = 3
	SourceBinormalB SourceRow = 4
	SourceTex0      SourceRow = 5
	SourceTex1      SourceRow = 6
	SourceTex2      SourceRow = 7
	SourceTex3      SourceRow = 8
	SourceTex4      SourceRow = 9
	SourceTex5      SourceRow = 10
	SourceTex6      SourceRow = 11
	SourceTex7      SourceRow = 12
)

// TexGenType selects a texture-coordinate generator's overall mode.
type TexGenType uint32

const (
	TexGenRegular      TexGenType = 0
	TexGenEmbossMap    TexGenType = 1
	TexGenColorSTRGBC0 TexGenType = 2
	TexGenColorSTRGBC1 TexGenType = 3
)

// TexGen decodes one XF texture-coordinate generator control register
// (TEX0-TEX7).
type TexGen struct {
	Raw uint32
}

func (t TexGen) EmbossLight() uint32  { return (t.Raw >> 15) & 0x7 }
func (t TexGen) EmbossSource() uint32 { return (t.Raw >> 12) & 0x7 }
func (t TexGen) SourceRow() SourceRow { return SourceRow((t.Raw >> 7) & 0x1f) }
func (t TexGen) Type() TexGenType     { return TexGenType((t.Raw >> 4) & 0x7) }

// PostTexGen decodes one XF post-transform (dual texture) register
// (DUALTEX0-7).
type PostTexGen struct {
	Raw uint32
}

// MatrixRow returns the base row of the dual-transform matrix this
// post-texture-generator stage uses (0-63, DualTexMatricesStart's own
// row indexing - matrix_select.go's PositionMatrixAddr documents the
// same row-index-to-word-address relationship for the position-matrix
// block).
func (p PostTexGen) MatrixRow() uint32 { return p.Raw & 0x3f }

// NormalizeBeforeTransform reports whether this stage's texture
// coordinate should be normalized before the dual transform is
// applied.
func (p PostTexGen) NormalizeBeforeTransform() bool { return p.Raw&0x100 != 0 }
