// cpregs.go decodes CP registers this project has a confirmed public
// layout for: matrix-index selection, the vertex descriptor (vcd.go),
// and the per-attribute array base/stride table. Addresses match
// public hardware notes (YAGCD chapter 5, "internal CP Registers").
// CP_VAT_REG_A/B/C (0x70-0x97, which would describe each attribute's
// exact component count/format/shift) are the one register group in
// that same source whose bit layout came through the extraction
// ambiguous enough that this project doesn't trust decoding it -
// vertex attribute values are still read using this project's own
// simplified fixed component sizes (vertexformat.go), the same
// simplification this package's vertex format already made before
// VCD-driven decoding existed.
package gpu

const (
	cpMatIdxRegA = 0x30
	cpMatIdxRegB = 0x40

	cpVCDLoBase = 0x50 // 0x50-0x57, 8 format slots
	cpVCDHiBase = 0x60 // 0x60-0x67

	// cpVATABase/B/C (0x70-0x97): raw-only, see package doc.

	cpArrayBaseBase   = 0xA0 // 0xA0-0xAF, one per attribute (see arrayXxx consts)
	cpArrayStrideBase = 0xB0 // 0xB0-0xBF
)

// Array register indices, in the order public hardware notes list
// ARRAY_BASE/ARRAY_STRIDE.
const (
	arrayPosition = 0
	arrayNormal   = 1
	arrayColor0   = 2
	arrayColor1   = 3
	arrayTex0     = 4
	// arrayTex1..arrayTex7 = 5..11; arrayGP0..arrayGP3 = 12..15
	// (general-purpose arrays this project doesn't use).
)

// applyCPRegisterWrite reacts to a LOAD_CP_REG write that already
// landed in cp.cpRegs, decoding it into the structured state
// vertexformat.go's dynamic vertex decoder consumes.
func (cp *CommandProcessor) applyCPRegisterWrite(reg byte, val uint32) {
	switch {
	case reg == cpMatIdxRegA:
		cp.matIdxA = MatIdxRegA(val)
		cp.bridgeMatrixSelectionToXF(val)
	case reg == cpMatIdxRegB:
		cp.matIdxB = MatIdxRegB(val)
		cp.bridgeMatrixSelection1ToXF(val)
	case reg >= cpVCDLoBase && reg < cpVCDLoBase+8:
		cp.vcdLo[reg-cpVCDLoBase] = VCDLo(val)
	case reg >= cpVCDHiBase && reg < cpVCDHiBase+8:
		cp.vcdHi[reg-cpVCDHiBase] = VCDHi(val)
	case reg >= cpArrayBaseBase && reg < cpArrayBaseBase+16:
		cp.arrayBase[reg-cpArrayBaseBase] = val & 0x03FFFFFF // bits 0-25
	case reg >= cpArrayStrideBase && reg < cpArrayStrideBase+16:
		cp.arrayStride[reg-cpArrayStrideBase] = byte(val) // bits 0-7
	}
}

// arrayEntry resolves one indexed attribute occurrence: index*stride
// bytes past the array's base address, read through memReader. A
// zero stride falls back to length (the attribute's own direct byte
// width), a reasonable default for the common case of tightly packed
// arrays real code actually uses.
func (cp *CommandProcessor) arrayEntry(arrayIdx int, index uint32, length int) []byte {
	if cp.memReader == nil {
		return make([]byte, length)
	}
	stride := uint32(cp.arrayStride[arrayIdx])
	if stride == 0 {
		stride = uint32(length)
	}
	return cp.memReader.ReadBytes(cp.arrayBase[arrayIdx]+index*stride, length)
}
