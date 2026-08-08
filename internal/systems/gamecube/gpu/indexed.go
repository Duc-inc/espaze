package gpu

// CP register addresses this project reserves for indexed vertex
// arrays. Real hardware controls this per-attribute (each of
// position/normal/texcoord/color can independently be NONE/DIRECT/
// INDEX8/INDEX16) via its vertex control descriptor; this project
// only supports switching position+normal+UV all together between
// fully direct (the existing fixed layout) and fully indexed with
// 16-bit indices, with color always direct either way - a real
// simplification of real hardware's much more flexible per-attribute
// scheme, in the same spirit as this project's fixed vertex layout.
const (
	cpVertexMode = 0x50 // 0 = direct (default), nonzero = indexed

	cpPosArrayBase      = 0x51 // byte address in main memory
	cpPosArrayStride    = 0x52 // bytes per entry (0 defaults to 6: 3x int16)
	cpNormalArrayBase   = 0x53
	cpNormalArrayStride = 0x54 // 0 defaults to 6: 3x int16
	cpUVArrayBase       = 0x55
	cpUVArrayStride     = 0x56 // 0 defaults to 4: 2x int16

	// indexedVertexBytes: 3x uint16 index (position, normal, UV) plus
	// a direct 4-byte color, per indexed vertex.
	indexedVertexBytes = 3*2 + 4
)

// isIndexed reports whether the current draw should read indices
// (indexedVertexBytes per vertex) instead of full direct vertex data.
func (cp *CommandProcessor) isIndexed() bool {
	return cp.cpRegs[cpVertexMode] != 0
}

func (cp *CommandProcessor) arrayEntry(base, stride, defaultStride, index, length uint32) []byte {
	if cp.memReader == nil {
		return make([]byte, length)
	}
	if stride == 0 {
		stride = defaultStride
	}
	return cp.memReader.ReadBytes(base+index*stride, int(length))
}

// decodeIndexedVertex reads three 16-bit array indices plus a direct
// color from b, and resolves the position/normal/UV entries they name
// via the current array base/stride CP registers and memReader.
func (cp *CommandProcessor) decodeIndexedVertex(b []byte) Vertex {
	posIdx := uint32(be16(b))
	normIdx := uint32(be16(b[2:]))
	uvIdx := uint32(be16(b[4:]))

	pos := cp.arrayEntry(cp.cpRegs[cpPosArrayBase], cp.cpRegs[cpPosArrayStride], 6, posIdx, 6)
	norm := cp.arrayEntry(cp.cpRegs[cpNormalArrayBase], cp.cpRegs[cpNormalArrayStride], 6, normIdx, 6)
	uv := cp.arrayEntry(cp.cpRegs[cpUVArrayBase], cp.cpRegs[cpUVArrayStride], 4, uvIdx, 4)

	return Vertex{
		X: int16(be16(pos)), Y: int16(be16(pos[2:])), Z: int16(be16(pos[4:])),
		NX: int16(be16(norm)), NY: int16(be16(norm[2:])), NZ: int16(be16(norm[4:])),
		U: int16(be16(uv)), V: int16(be16(uv[2:])),
		R: b[6], G: b[7], B: b[8], A: b[9],
	}
}
