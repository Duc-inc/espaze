package gpu

import (
	"math"
	"testing"
)

func loadCPRegBytes(reg byte, val uint32) []byte {
	return []byte{cmdLoadCPReg, reg, byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val)}
}

func loadVCDLoBytes(format byte, word uint32) []byte {
	return loadCPRegBytes(cpVCDLoBase+format, word)
}
func loadVCDHiBytes(format byte, word uint32) []byte {
	return loadCPRegBytes(cpVCDHiBase+format, word)
}
func loadArrayBaseBytes(attr byte, addr uint32) []byte {
	return loadCPRegBytes(cpArrayBaseBase+attr, addr)
}
func loadArrayStrideBytes(attr byte, stride uint32) []byte {
	return loadCPRegBytes(cpArrayStrideBase+attr, stride)
}

func TestVCDPositionAsIndex16ReadsFromRealArrayRegisters(t *testing.T) {
	cp := New()
	mem := &fakeMemory{data: make([]byte, 0x200)}
	// Position array at 0x100, stride 6: entry 1 = (10,20,30).
	copy(mem.data[0x106:], []byte{0, 10, 0, 20, 0, 30})
	cp.SetMemoryReader(mem)

	var stream []byte
	// VCD_LO: Position=Index16(3), Normal=Direct(1), Color0=Direct(1).
	stream = append(stream, loadVCDLoBytes(0, uint32(AttrIndex16)<<9|uint32(AttrDirect)<<11|uint32(AttrDirect)<<13)...)
	// VCD_HI: Tex0=Direct(1).
	stream = append(stream, loadVCDHiBytes(0, uint32(AttrDirect))...)
	stream = append(stream, loadArrayBaseBytes(arrayPosition, 0x100)...)
	stream = append(stream, loadArrayStrideBytes(arrayPosition, 6)...)

	stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
	// Each vertex: posIdx(2) + normal(6) + color(4) + uv(4) = 16 bytes.
	vertex := func(idx uint16) []byte {
		return []byte{
			byte(idx >> 8), byte(idx),
			0, 0, 0, 0, 0, 0, // normal
			255, 255, 255, 255, // color
			0, 0, 0, 0, // uv
		}
	}
	stream = append(stream, vertex(1)...)
	stream = append(stream, vertex(1)...)
	stream = append(stream, vertex(1)...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V0.X != 10 || tris[0].V0.Y != 20 || tris[0].V0.Z != 30 {
		t.Fatalf("V0 position = (%d,%d,%d), want (10,20,30) resolved via real VCD/ARRAY registers",
			tris[0].V0.X, tris[0].V0.Y, tris[0].V0.Z)
	}
}

func TestVCDAbsentTexCoordLeavesUVZeroAndShrinksStreamWidth(t *testing.T) {
	cp := New()
	var stream []byte
	// VCD_LO: Position/Normal/Color0 all Direct; VCD_HI stays 0 (Tex0 absent).
	stream = append(stream, loadVCDLoBytes(0, uint32(AttrDirect)<<9|uint32(AttrDirect)<<11|uint32(AttrDirect)<<13)...)

	stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
	// Each vertex: position(6) + normal(6) + color(4) = 16 bytes, no UV.
	vertex := []byte{
		0, 5, 0, 0, 0, 0, // position (5,0,0)
		0, 0, 0, 0, 0, 0, // normal
		200, 100, 50, 255, // color
	}
	stream = append(stream, vertex...)
	stream = append(stream, vertex...)
	stream = append(stream, vertex...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V0.X != 5 {
		t.Fatalf("V0.X = %d, want 5 (stream stayed in sync with no UV present)", tris[0].V0.X)
	}
	if tris[0].V0.U != 0 || tris[0].V0.V != 0 {
		t.Fatalf("V0 UV = (%d,%d), want (0,0): tex0 wasn't present in this vertex descriptor", tris[0].V0.U, tris[0].V0.V)
	}
}

func TestMatIdxRegABridgesToRealXFMatrixSelection(t *testing.T) {
	cp := New()
	translate := []float32{
		1, 0, 0, 30,
		0, 1, 0, 0,
		0, 0, 1, 0,
	}
	var stream []byte
	for i, f := range translate {
		stream = append(stream, loadXFRegBytes(uint16(4+i), math.Float32bits(f))...)
	}
	// MATIDX_REG_A: PosNormalIndex=1 -> row 1 (word address 4).
	stream = append(stream, loadCPRegBytes(cpMatIdxRegA, 1)...)

	stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
	v := Vertex{X: 0, Y: 0}
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V0.X != 30 {
		t.Fatalf("V0.X = %d, want 30 (CP MATIDX_REG_A should drive the real XF matrix selection)", tris[0].V0.X)
	}
}
