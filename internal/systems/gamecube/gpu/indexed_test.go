package gpu

import "testing"

func loadCPRegBytes(reg byte, val uint32) []byte {
	return []byte{cmdLoadCPReg, reg, byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val)}
}

func TestIndexedVertexModeReadsFromArraysInMemory(t *testing.T) {
	cp := New()
	mem := &fakeMemory{data: make([]byte, 0x200)}
	// Position array at 0x100: entry 0 = (0,0,0), entry 1 = (10,20,30).
	copy(mem.data[0x100:], []byte{0, 0, 0, 0, 0, 0})
	copy(mem.data[0x106:], []byte{0, 10, 0, 20, 0, 30})
	// UV array at 0x140: entry 0 = (0,0), entry 1 = (5,7).
	copy(mem.data[0x144:], []byte{0, 5, 0, 7})
	cp.SetMemoryReader(mem)

	var stream []byte
	stream = append(stream, loadCPRegBytes(cpVertexMode, 1)...)
	stream = append(stream, loadCPRegBytes(cpPosArrayBase, 0x100)...)
	stream = append(stream, loadCPRegBytes(cpUVArrayBase, 0x140)...)

	stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
	// Each indexed vertex: posIdx(2) normIdx(2) uvIdx(2) + RGBA(4).
	v0 := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 255, 255, 255, 255}
	v1 := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 255, 255, 255, 255}
	v2 := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 255, 255, 255, 255}
	stream = append(stream, v0...)
	stream = append(stream, v1...)
	stream = append(stream, v2...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V1.X != 10 || tris[0].V1.Y != 20 || tris[0].V1.Z != 30 {
		t.Fatalf("V1 position = (%d,%d,%d), want (10,20,30) from array index 1",
			tris[0].V1.X, tris[0].V1.Y, tris[0].V1.Z)
	}
	if tris[0].V1.U != 5 || tris[0].V1.V != 7 {
		t.Fatalf("V1 UV = (%d,%d), want (5,7) from array index 1", tris[0].V1.U, tris[0].V1.V)
	}
	if tris[0].V0.X != 0 || tris[0].V0.Y != 0 || tris[0].V0.Z != 0 {
		t.Fatalf("V0 position = (%d,%d,%d), want (0,0,0) from array index 0",
			tris[0].V0.X, tris[0].V0.Y, tris[0].V0.Z)
	}
}

func TestDirectVertexModeIsUnaffectedByArrayRegisters(t *testing.T) {
	// cpVertexMode defaults to 0 (direct): existing fixed-layout
	// decoding must still work exactly as before, ignoring any array
	// registers that happen to be set.
	cp := New()
	stream := []byte{cmdDrawTriangles, 0x00, 0x03}
	v := Vertex{X: 1, Y: 2, Z: 3, R: 255, G: 255, B: 255, A: 255}
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V0.X != 1 || tris[0].V0.Y != 2 || tris[0].V0.Z != 3 {
		t.Fatalf("V0 position = (%d,%d,%d), want (1,2,3)", tris[0].V0.X, tris[0].V0.Y, tris[0].V0.Z)
	}
}

func TestIndexedVertexWithoutMemoryReaderReadsZeroedEntries(t *testing.T) {
	cp := New()
	var stream []byte
	stream = append(stream, loadCPRegBytes(cpVertexMode, 1)...)
	stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
	v := []byte{0x00, 0x05, 0x00, 0x05, 0x00, 0x05, 255, 255, 255, 255}
	stream = append(stream, v...)
	stream = append(stream, v...)
	stream = append(stream, v...)
	cp.Execute(stream) // must not panic without a memory reader

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V0.X != 0 || tris[0].V0.Y != 0 || tris[0].V0.Z != 0 {
		t.Fatalf("V0 position = (%d,%d,%d), want (0,0,0) with no memory reader", tris[0].V0.X, tris[0].V0.Y, tris[0].V0.Z)
	}
}
