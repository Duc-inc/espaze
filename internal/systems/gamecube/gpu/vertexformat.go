// vertexformat.go decodes a real, per-draw-call configurable vertex
// layout from CP's vertex descriptor (vcd.go), replacing this
// package's previous single fixed layout. Each attribute (position,
// normal, color0, color1, texcoord0-7) can independently be absent,
// carried directly in the command stream, or supplied as an 8/16-bit
// index into a real ARRAY_BASE/ARRAY_STRIDE-addressed array in main
// memory (cpregs.go) - the actual mechanism real games use to share
// vertex data across draw calls, rather than this project's earlier
// single "direct vs. indexed, position+normal+UV all together" stand-
// in. This project's Vertex type still only has one color and one
// texcoord field, so Color1 and TexCoord1-7 are read (to keep the
// command stream in sync) but discarded. A per-vertex position/normal
// matrix-index override is read into Vertex.MatrixIndex and applied by
// decode.go's transformVertex - real hardware's per-vertex skinning.
package gpu

// defaultVCDLo/defaultVCDHi describe this project's original fixed
// vertex layout (position+normal+color0+texcoord0, all direct) - used
// until a command stream writes real VCD_LO/VCD_HI registers, so
// behavior is unchanged for any caller that never configures them.
var (
	defaultVCDLo = VCDLo(uint32(AttrDirect)<<9 | uint32(AttrDirect)<<11 | uint32(AttrDirect)<<13)
	defaultVCDHi = VCDHi(uint32(AttrDirect))
)

// vertexDescriptor pairs one draw call's active VCD_LO/VCD_HI.
type vertexDescriptor struct {
	lo VCDLo
	hi VCDHi
}

// currentVCD returns format slot 0's descriptor, or this project's
// default fixed layout if the command stream hasn't configured VCD_LO/
// VCD_HI at all yet. Real hardware supports 8 format slots and
// presumably selects among them per draw call; this project always
// uses slot 0, a simplification this package's vertex format already
// made in spirit before VCD-driven decoding existed.
func (cp *CommandProcessor) currentVCD() vertexDescriptor {
	lo, hi := cp.vcdLo[0], cp.vcdHi[0]
	if lo == 0 && hi == 0 {
		return vertexDescriptor{defaultVCDLo, defaultVCDHi}
	}
	return vertexDescriptor{lo, hi}
}

// attrStreamBytes returns how many command-stream bytes one occurrence
// of enc consumes: directBytes for AttrDirect (this project's own
// simplified per-component size, see cpregs.go's package doc for why
// the real VAT format registers aren't decoded), 1 for AttrIndex8, 2
// for AttrIndex16, 0 for AttrNone.
func attrStreamBytes(enc AttrEncoding, directBytes int) int {
	switch enc {
	case AttrDirect:
		return directBytes
	case AttrIndex8:
		return 1
	case AttrIndex16:
		return 2
	default: // AttrNone
		return 0
	}
}

// byteWidth returns one vertex's total command-stream size under this
// descriptor.
func (vcd vertexDescriptor) byteWidth() int {
	n := 0
	if vcd.lo.PosNormalMatrixIdxPresent() {
		n++
	}
	for i := 0; i < 8; i++ {
		if vcd.lo.TexMatrixIdxPresent(i) {
			n++
		}
	}
	n += attrStreamBytes(vcd.lo.Position(), 6)
	n += attrStreamBytes(vcd.lo.Normal(), 6)
	n += attrStreamBytes(vcd.lo.Color0(), 4)
	n += attrStreamBytes(vcd.lo.Color1(), 4)
	for i := 0; i < 8; i++ {
		n += attrStreamBytes(vcd.hi.Tex(i), 4)
	}
	return n
}

// readAttr returns one attribute occurrence's directBytes-long value:
// read straight from b if enc is AttrDirect, resolved through
// arrayIdx's ARRAY_BASE/ARRAY_STRIDE if enc is an index encoding, or
// zeroed if enc is AttrNone. cursor is advanced by however many stream
// bytes this occurrence consumed. Callers are expected to have already
// bounds-checked b against the descriptor's byteWidth(), so the
// zeroed-fallback paths here are a safety net, not the normal case.
func (cp *CommandProcessor) readAttr(enc AttrEncoding, arrayIdx int, b []byte, cursor *int, directBytes int) []byte {
	switch enc {
	case AttrDirect:
		if *cursor+directBytes > len(b) {
			return make([]byte, directBytes)
		}
		data := b[*cursor : *cursor+directBytes]
		*cursor += directBytes
		return data
	case AttrIndex8, AttrIndex16:
		width := attrStreamBytes(enc, directBytes)
		if *cursor+width > len(b) {
			return make([]byte, directBytes)
		}
		var index uint32
		if enc == AttrIndex8 {
			index = uint32(b[*cursor])
		} else {
			index = uint32(be16(b[*cursor:]))
		}
		*cursor += width
		return cp.arrayEntry(arrayIdx, index, directBytes)
	default: // AttrNone
		return make([]byte, directBytes)
	}
}

// decodeDynamicVertex reads one vertex from b according to vcd. The
// caller is expected to have already bounds-checked b against
// vcd.byteWidth().
func (cp *CommandProcessor) decodeDynamicVertex(vcd vertexDescriptor, b []byte) Vertex {
	var v Vertex
	cursor := 0

	// Per-vertex position/normal matrix-index override: real hardware's
	// per-vertex skinning mechanism. transformVertex (decode.go) applies
	// it in place of xf.RegMatrixSelection0's GeometryIndex when
	// present.
	if vcd.lo.PosNormalMatrixIdxPresent() {
		if cursor < len(b) {
			v.HasMatrixIndexOverride = true
			v.MatrixIndex = b[cursor]
		}
		cursor++
	}
	for i := 0; i < 8; i++ {
		if vcd.lo.TexMatrixIdxPresent(i) {
			cursor++
		}
	}

	pos := cp.readAttr(vcd.lo.Position(), arrayPosition, b, &cursor, 6)
	v.X, v.Y, v.Z = int16(be16(pos)), int16(be16(pos[2:])), int16(be16(pos[4:]))

	nrm := cp.readAttr(vcd.lo.Normal(), arrayNormal, b, &cursor, 6)
	v.NX, v.NY, v.NZ = int16(be16(nrm)), int16(be16(nrm[2:])), int16(be16(nrm[4:]))

	col0 := cp.readAttr(vcd.lo.Color0(), arrayColor0, b, &cursor, 4)
	v.R, v.G, v.B, v.A = col0[0], col0[1], col0[2], col0[3]

	cp.readAttr(vcd.lo.Color1(), arrayColor1, b, &cursor, 4) // discarded, see package doc

	tex0 := cp.readAttr(vcd.hi.Tex(0), arrayTex0, b, &cursor, 4)
	v.U, v.V = int16(be16(tex0)), int16(be16(tex0[2:]))

	for i := 1; i < 8; i++ {
		cp.readAttr(vcd.hi.Tex(i), arrayTex0+i, b, &cursor, 4) // discarded, see package doc
	}

	return v
}
