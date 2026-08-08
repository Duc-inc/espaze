package gpu

import "testing"

func vertexBytesOf(v Vertex) []byte {
	return []byte{
		byte(v.X >> 8), byte(v.X), byte(v.Y >> 8), byte(v.Y), byte(v.Z >> 8), byte(v.Z),
		v.R, v.G, v.B, v.A,
	}
}

func TestNopIsSkipped(t *testing.T) {
	cp := New()
	cp.Execute([]byte{cmdNop, cmdNop})
	if len(cp.DrainTriangles()) != 0 {
		t.Fatal("expected no triangles from NOPs")
	}
}

func TestLoadCPRegStoresValue(t *testing.T) {
	cp := New()
	cp.Execute([]byte{cmdLoadCPReg, 0x50, 0x00, 0x00, 0x00, 0x2A})
	if cp.cpRegs[0x50] != 0x2A {
		t.Fatalf("cpRegs[0x50] = %#08x, want 0x2A", cp.cpRegs[0x50])
	}
}

func TestLoadBPRegPacksAddressAndData(t *testing.T) {
	cp := New()
	// Address 0x40, data 0x001234 packed into one 4-byte word.
	cp.Execute([]byte{cmdLoadBPReg, 0x40, 0x00, 0x12, 0x34})
	if cp.bpRegs[0x40] != 0x001234 {
		t.Fatalf("bpRegs[0x40] = %#06x, want 0x001234", cp.bpRegs[0x40])
	}
}

func TestDrawTrianglesProducesOneTrianglePerThreeVertices(t *testing.T) {
	cp := New()
	v0 := Vertex{X: 0, Y: 0, Z: 0, R: 255}
	v1 := Vertex{X: 10, Y: 0, Z: 0, G: 255}
	v2 := Vertex{X: 5, Y: 10, Z: 0, B: 255}

	stream := []byte{cmdDrawTriangles, 0x00, 0x03}
	stream = append(stream, vertexBytesOf(v0)...)
	stream = append(stream, vertexBytesOf(v1)...)
	stream = append(stream, vertexBytesOf(v2)...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V0 != v0 || tris[0].V1 != v1 || tris[0].V2 != v2 {
		t.Fatalf("triangle vertices = %+v, want %v/%v/%v", tris[0], v0, v1, v2)
	}
}

func TestDrawQuadsProducesTwoTrianglesPerFourVertices(t *testing.T) {
	cp := New()
	verts := []Vertex{
		{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10},
	}
	stream := []byte{cmdDrawQuads, 0x00, 0x04}
	for _, v := range verts {
		stream = append(stream, vertexBytesOf(v)...)
	}
	cp.Execute(stream)

	if len(cp.DrainTriangles()) != 2 {
		t.Fatalf("got %d triangles, want 2", len(cp.DrainTriangles()))
	}
}

func TestDrawTriangleFanExpandsCorrectly(t *testing.T) {
	cp := New()
	verts := []Vertex{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}}
	stream := []byte{cmdDrawTriFan, 0x00, 0x04}
	for _, v := range verts {
		stream = append(stream, vertexBytesOf(v)...)
	}
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 2 {
		t.Fatalf("got %d triangles, want 2 (a 4-vertex fan)", len(tris))
	}
	if tris[0].V0 != verts[0] {
		t.Fatal("expected every fan triangle to share vertex 0")
	}
}
