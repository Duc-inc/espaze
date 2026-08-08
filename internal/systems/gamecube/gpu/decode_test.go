package gpu

import (
	"math"
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/xf"
)

func vertexBytesOf(v Vertex) []byte {
	return []byte{
		byte(v.X >> 8), byte(v.X), byte(v.Y >> 8), byte(v.Y), byte(v.Z >> 8), byte(v.Z),
		byte(v.NX >> 8), byte(v.NX), byte(v.NY >> 8), byte(v.NY), byte(v.NZ >> 8), byte(v.NZ),
		byte(v.U >> 8), byte(v.U), byte(v.V >> 8), byte(v.V),
		v.R, v.G, v.B, v.A,
	}
}

// loadXFRegBytes builds one LOAD_XF_REG command uploading a single
// 32-bit word at addr.
func loadXFRegBytes(addr uint16, word uint32) []byte {
	return []byte{
		cmdLoadXFReg,
		byte(addr >> 8), byte(addr),
		byte(word >> 24), byte(word >> 16), byte(word >> 8), byte(word),
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

func TestLoadXFRegUploadsMatrixThatTransformsVertices(t *testing.T) {
	cp := New()

	// Overwrite the default identity position matrix (address 0) with
	// one that translates by (10, 0, 0), one LOAD_XF_REG word at a
	// time - exactly how a real game uploads a modelview matrix.
	translate := []float32{
		1, 0, 0, 10,
		0, 1, 0, 0,
		0, 0, 1, 0,
	}
	stream := []byte{}
	for i, f := range translate {
		stream = append(stream, loadXFRegBytes(uint16(i), math.Float32bits(f))...)
	}

	stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
	verts := []Vertex{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}
	for _, v := range verts {
		stream = append(stream, vertexBytesOf(v)...)
	}

	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V0.X != 10 {
		t.Fatalf("V0.X = %d, want 10 (translation from the uploaded XF matrix)", tris[0].V0.X)
	}
	if tris[0].V1.X != 11 {
		t.Fatalf("V1.X = %d, want 11", tris[0].V1.X)
	}
}

func TestLightingLeavesColorUnchangedWithFullyFacingLightAndNoAmbient(t *testing.T) {
	cp := New()
	// Give the normal matrix its own address so it doesn't overlap
	// the identity position matrix New() already wrote at address 0,
	// and make it identity too.
	cp.xfMemory.WriteNormalMatrix(100, xf.NormalMatrix{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	})
	cp.xfRegisters.NormalMatrixIndex = 100

	cp.SetAmbient(xf.Ambient{})
	cp.SetLight(0, xf.Light{
		Position: xf.Vec3{X: 0, Y: 0, Z: 10},
		Color:    xf.LightColor{R: 1, G: 1, B: 1},
		Enabled:  true,
	})

	v := Vertex{X: 0, Y: 0, Z: 0, NZ: 1, R: 255, G: 255, B: 255, A: 255}
	stream := []byte{cmdDrawTriangles, 0x00, 0x03}
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V0.R != 255 || tris[0].V0.G != 255 || tris[0].V0.B != 255 {
		t.Fatalf("color = (%d,%d,%d), want (255,255,255): a light directly along the normal with white material should pass through unchanged",
			tris[0].V0.R, tris[0].V0.G, tris[0].V0.B)
	}
}

func TestLightingDarkensSurfaceFacingAwayFromLight(t *testing.T) {
	cp := New()
	cp.xfMemory.WriteNormalMatrix(100, xf.NormalMatrix{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	})
	cp.xfRegisters.NormalMatrixIndex = 100

	cp.SetAmbient(xf.Ambient{}) // no ambient term either
	cp.SetLight(0, xf.Light{
		Position: xf.Vec3{X: 0, Y: 0, Z: -10}, // behind the surface
		Color:    xf.LightColor{R: 1, G: 1, B: 1},
		Enabled:  true,
	})

	v := Vertex{X: 0, Y: 0, Z: 0, NZ: 1, R: 255, G: 255, B: 255, A: 255}
	stream := []byte{cmdDrawTriangles, 0x00, 0x03}
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V0.R != 0 || tris[0].V0.G != 0 || tris[0].V0.B != 0 {
		t.Fatalf("color = (%d,%d,%d), want (0,0,0): a light behind the surface with no ambient should leave it unlit",
			tris[0].V0.R, tris[0].V0.G, tris[0].V0.B)
	}
}
