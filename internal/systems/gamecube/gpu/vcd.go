// vcd.go decodes CP's Vertex Control Descriptor (VCD_LO/VCD_HI) and
// matrix-index registers (MATIDX_REG_A/B) - real hardware's own way of
// describing which attributes a vertex carries and how (absent/
// direct/8-bit index/16-bit index), instead of this project's
// previous single fixed layout. Bit positions below come from a
// public hardware register reference (YAGCD chapter 5, "internal CP
// Registers"), extracted and cross-checked directly from that source.
// The companion Vertex Attribute Table registers (CP 0x70-0x97, which
// would describe each attribute's exact component count/format/shift)
// have an ambiguous bit layout in that same source and are
// deliberately not decoded here - see cpregs.go's own doc comment.
package gpu

// AttrEncoding is how one vertex attribute's data is supplied in the
// command stream.
type AttrEncoding byte

const (
	AttrNone    AttrEncoding = 0
	AttrDirect  AttrEncoding = 1
	AttrIndex8  AttrEncoding = 2
	AttrIndex16 AttrEncoding = 3
)

// VCDLo decodes CP's Vertex Descriptor low register (0x50-0x57, one
// per format slot): per-attribute presence flags for the position/
// normal and per-texture matrix-index overrides, plus 2-bit encodings
// for position/normal/color0/color1.
type VCDLo uint32

// PosNormalMatrixIdxPresent reports whether the vertex stream carries
// a per-vertex position/normal matrix index override (assumed to be a
// single index byte when present - the source describes this as a
// presence flag, not a 2-bit encoding, so unlike Position/Normal/
// Color0/Color1 there's no direct/index8/index16 distinction to make).
func (v VCDLo) PosNormalMatrixIdxPresent() bool { return uint32(v)&1 != 0 }

// TexMatrixIdxPresent reports the same presence flag for texture
// matrix n (0-7).
func (v VCDLo) TexMatrixIdxPresent(n int) bool { return uint32(v)&(2<<uint(n)) != 0 }

func (v VCDLo) Position() AttrEncoding { return AttrEncoding((uint32(v) >> 9) & 0x3) }
func (v VCDLo) Normal() AttrEncoding   { return AttrEncoding((uint32(v) >> 11) & 0x3) }
func (v VCDLo) Color0() AttrEncoding   { return AttrEncoding((uint32(v) >> 13) & 0x3) }
func (v VCDLo) Color1() AttrEncoding   { return AttrEncoding((uint32(v) >> 15) & 0x3) }

// VCDHi decodes CP's Vertex Descriptor high register (0x60-0x67, one
// per format slot): 2-bit encodings for texture coordinates 0-7.
type VCDHi uint32

func (v VCDHi) Tex(n int) AttrEncoding { return AttrEncoding((uint32(v) >> uint(2*n)) & 0x3) }

// MatIdxRegA decodes CP's MATIDX_REG_A (0x30): the position/normal
// matrix index plus texture matrix indices 0-3, each a 6-bit field -
// the same bit layout xf.MatrixSelection0 uses for the equivalent XF
// register, which is why bind.go can forward this register's raw
// value straight into xf.RegMatrixSelection0.
type MatIdxRegA uint32

func (m MatIdxRegA) PosNormalIndex() uint32 { return uint32(m) & 0x3f }
func (m MatIdxRegA) Tex0Index() uint32      { return (uint32(m) >> 6) & 0x3f }
func (m MatIdxRegA) Tex1Index() uint32      { return (uint32(m) >> 12) & 0x3f }
func (m MatIdxRegA) Tex2Index() uint32      { return (uint32(m) >> 18) & 0x3f }
func (m MatIdxRegA) Tex3Index() uint32      { return (uint32(m) >> 24) & 0x3f }

// MatIdxRegB decodes CP's MATIDX_REG_B (0x40): texture matrix indices
// 4-7, each a 6-bit field - matches xf.MatrixSelection1's layout, same
// reasoning as MatIdxRegA.
type MatIdxRegB uint32

func (m MatIdxRegB) Tex4Index() uint32 { return uint32(m) & 0x3f }
func (m MatIdxRegB) Tex5Index() uint32 { return (uint32(m) >> 6) & 0x3f }
func (m MatIdxRegB) Tex6Index() uint32 { return (uint32(m) >> 12) & 0x3f }
func (m MatIdxRegB) Tex7Index() uint32 { return (uint32(m) >> 18) & 0x3f }
