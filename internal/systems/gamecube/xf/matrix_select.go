// matrix_select.go decodes XF's matrix-selection registers (RegMatrix
// Selection0/1, registers.go) - public hardware notes describe each as
// four or five packed 6-bit matrix-index fields.
package xf

// MatrixSelection0 packs the geometry (position) matrix index plus
// the first four texture-matrix indices, each a 6-bit field.
type MatrixSelection0 uint32

func (m MatrixSelection0) GeometryIndex() uint32 { return uint32(m) & 0x3f }
func (m MatrixSelection0) Texture0Index() uint32 { return (uint32(m) >> 6) & 0x3f }
func (m MatrixSelection0) Texture1Index() uint32 { return (uint32(m) >> 12) & 0x3f }
func (m MatrixSelection0) Texture2Index() uint32 { return (uint32(m) >> 18) & 0x3f }
func (m MatrixSelection0) Texture3Index() uint32 { return (uint32(m) >> 24) & 0x3f }

// MatrixSelection1 packs texture-matrix indices 4-7, each a 6-bit
// field.
type MatrixSelection1 uint32

func (m MatrixSelection1) Texture4Index() uint32 { return uint32(m) & 0x3f }
func (m MatrixSelection1) Texture5Index() uint32 { return (uint32(m) >> 6) & 0x3f }
func (m MatrixSelection1) Texture6Index() uint32 { return (uint32(m) >> 12) & 0x3f }
func (m MatrixSelection1) Texture7Index() uint32 { return (uint32(m) >> 18) & 0x3f }

// PositionMatrixAddr returns the XF-memory word address of the
// position matrix GeometryIndex selects. The position-matrix block
// (PosMatricesStart..PosMatricesEnd, memory.go) holds 256 words as 64
// rows of 4, so a 6-bit index (0-63) multiplies by 4 to become a word
// offset - the only arithmetic that makes every index address a
// distinct, in-range row.
func (r Registers) PositionMatrixAddr() uint16 {
	return PosMatricesStart + uint16(r.MatrixSelection0.GeometryIndex()*4)
}

// TextureMatrixAddr returns the XF-memory word address of texture
// generator n's (0-7) selected matrix. Public hardware notes describe
// the texture-matrix indices (Texture0-7Index) as 6-bit fields with
// the same row shape as the position-matrix block, so this addresses
// the same 0x000-0x0ff block PositionMatrixAddr does, just at a
// different (typically higher) row - GameCube XF shares one matrix
// RAM between position and texture matrices rather than giving
// texture matrices a separate block. An out-of-range n returns
// PosMatricesStart (row 0).
func (r Registers) TextureMatrixAddr(n int) uint16 {
	var idx uint32
	switch n {
	case 0:
		idx = r.MatrixSelection0.Texture0Index()
	case 1:
		idx = r.MatrixSelection0.Texture1Index()
	case 2:
		idx = r.MatrixSelection0.Texture2Index()
	case 3:
		idx = r.MatrixSelection0.Texture3Index()
	case 4:
		idx = r.MatrixSelection1.Texture4Index()
	case 5:
		idx = r.MatrixSelection1.Texture5Index()
	case 6:
		idx = r.MatrixSelection1.Texture6Index()
	case 7:
		idx = r.MatrixSelection1.Texture7Index()
	}
	return PosMatricesStart + uint16(idx*4)
}

// NormalMatrixAddr returns the XF-memory word address of the normal
// matrix GeometryIndex selects. Public hardware notes on
// RegMatrixSelection0 state position and normal matrices are stored
// separately, but share the same index: "if index A is used for the
// position, then index A needs to be used for the normal as well" -
// so this multiplies the same GeometryIndex by 3 (the normal-matrix
// block's row width, NormalMatricesStart..NormalMatricesEnd holding 96
// words as 32 rows of 3) rather than tracking a separate field. An
// index past 31 addresses past the documented normal-matrix block;
// real hardware's behavior there isn't covered by this project's
// source, but Memory's own bounds check keeps it from reading out of
// the whole XF memory array.
func (r Registers) NormalMatrixAddr() uint16 {
	return NormalMatricesStart + uint16(r.MatrixSelection0.GeometryIndex()*3)
}
