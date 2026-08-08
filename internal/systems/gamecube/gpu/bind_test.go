package gpu

import (
	"math"
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/texture"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/xf"
)

type fakeMemory struct{ data []byte }

func (f *fakeMemory) ReadBytes(addr uint32, length int) []byte {
	out := make([]byte, length)
	if int(addr) >= len(f.data) {
		return out
	}
	copy(out, f.data[addr:])
	return out
}

// loadBPRegBytes builds one LOAD_BP_REG command: opcode, then a
// 4-byte word packing the register address in the top byte and a
// 24-bit data value in the rest, matching cmdLoadBPReg's decode.
func loadBPRegBytes(reg byte, data uint32) []byte {
	return []byte{cmdLoadBPReg, reg, byte(data >> 16), byte(data >> 8), byte(data)}
}

func TestBPRegistersBindTextureFromMemory(t *testing.T) {
	cp := New()
	mem := &fakeMemory{data: make([]byte, 0x200)}
	copy(mem.data[0x100:], []byte{0, 0, 255, 255}) // one blue RGBA8 texel
	cp.SetMemoryReader(mem)

	var stream []byte
	stream = append(stream, loadBPRegBytes(bpTexAddr, 0x100)...)
	stream = append(stream, loadBPRegBytes(bpTexFormat, uint32(texture.FormatRGBA8))...)
	stream = append(stream, loadBPRegBytes(bpTexWidth, 1)...)
	stream = append(stream, loadBPRegBytes(bpTexHeight, 1)...)
	stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
	v := Vertex{X: 0, Y: 0, R: 255, G: 255, B: 255, A: 255}
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].Textures[0] == nil {
		t.Fatal("expected a texture to be bound from BP register writes")
	}
	sampled := tris[0].Textures[0].Sample(0, 0)
	if sampled.R != 0 || sampled.G != 0 || sampled.B != 255 {
		t.Fatalf("sampled = %+v, want blue (0,0,255)", sampled)
	}
}

func TestBPTevOpSwitchesFromDefaultModulateToReplace(t *testing.T) {
	cp := New()
	mem := &fakeMemory{data: make([]byte, 0x200)}
	copy(mem.data[0x100:], []byte{255, 0, 0, 255}) // one red RGBA8 texel
	cp.SetMemoryReader(mem)

	var bind []byte
	bind = append(bind, loadBPRegBytes(bpTexAddr, 0x100)...)
	bind = append(bind, loadBPRegBytes(bpTexFormat, uint32(texture.FormatRGBA8))...)
	bind = append(bind, loadBPRegBytes(bpTexWidth, 1)...)
	bind = append(bind, loadBPRegBytes(bpTexHeight, 1)...)
	cp.Execute(bind)

	draw := func() Triangle {
		var stream []byte
		stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
		v0 := Vertex{X: 0, Y: 0, G: 255, A: 255}
		v1 := Vertex{X: 64, Y: 0, G: 255, A: 255}
		v2 := Vertex{X: 32, Y: 64, G: 255, A: 255}
		stream = append(stream, vertexBytesOf(v0)...)
		stream = append(stream, vertexBytesOf(v1)...)
		stream = append(stream, vertexBytesOf(v2)...)
		cp.Execute(stream)
		tris := cp.DrainTriangles()
		if len(tris) != 1 {
			t.Fatalf("got %d triangles, want 1", len(tris))
		}
		return tris[0]
	}

	// Default (Modulate): red texel * green vertex = black.
	f := NewFramebuffer(64, 64)
	f.Clear()
	f.DrawTriangle(draw())
	fb := f.FrameBuffer()
	i := (20*fb.Width + 32) * 4
	if fb.Pixels[i] != 0 || fb.Pixels[i+1] != 0 {
		t.Fatalf("modulate pixel = (%d,%d,%d), want black (red texel * green vertex)", fb.Pixels[i], fb.Pixels[i+1], fb.Pixels[i+2])
	}

	// Switch to Replace via BP register: the texel wins outright.
	cp.Execute(loadBPRegBytes(bpTevOp, uint32(TEVReplace)))
	f2 := NewFramebuffer(64, 64)
	f2.Clear()
	f2.DrawTriangle(draw())
	fb2 := f2.FrameBuffer()
	if fb2.Pixels[i] != 255 || fb2.Pixels[i+1] != 0 {
		t.Fatalf("replace pixel = (%d,%d,%d), want red (texel wins outright)", fb2.Pixels[i], fb2.Pixels[i+1], fb2.Pixels[i+2])
	}
}

func TestBPTextureBindWithoutMemoryReaderStaysNil(t *testing.T) {
	cp := New() // no SetMemoryReader call

	var stream []byte
	stream = append(stream, loadBPRegBytes(bpTexAddr, 0x100)...)
	stream = append(stream, loadBPRegBytes(bpTexFormat, uint32(texture.FormatRGBA8))...)
	stream = append(stream, loadBPRegBytes(bpTexWidth, 1)...)
	stream = append(stream, loadBPRegBytes(bpTexHeight, 1)...)
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
	if tris[0].Textures[0] != nil {
		t.Fatal("expected no texture bound without a memory reader")
	}
}

func TestXFRegistersBindAmbient(t *testing.T) {
	cp := New()
	var stream []byte
	stream = append(stream, loadXFRegBytes(xfRegAmbientR, math.Float32bits(0.5))...)
	stream = append(stream, loadXFRegBytes(xfRegAmbientG, math.Float32bits(0.25))...)
	stream = append(stream, loadXFRegBytes(xfRegAmbientB, math.Float32bits(0))...)
	cp.Execute(stream)

	want := xf.LightColor{R: 0.5, G: 0.25, B: 0}
	if cp.ambient.Color != want {
		t.Fatalf("ambient = %+v, want %+v", cp.ambient.Color, want)
	}
}

func TestXFRegistersBindLightAndAffectRendering(t *testing.T) {
	cp := New()
	cp.xfMemory.WriteNormalMatrix(100, xf.NormalMatrix{
		{1, 0, 0}, {0, 1, 0}, {0, 0, 1},
	})
	cp.xfRegisters.NormalMatrixIndex = 100

	var stream []byte
	stream = append(stream, loadXFRegBytes(xfRegAmbientR, math.Float32bits(0))...)
	stream = append(stream, loadXFRegBytes(xfRegAmbientG, math.Float32bits(0))...)
	stream = append(stream, loadXFRegBytes(xfRegAmbientB, math.Float32bits(0))...)
	base := uint16(xfRegLightBase)
	stream = append(stream, loadXFRegBytes(base+0, math.Float32bits(0))...) // pos.x
	stream = append(stream, loadXFRegBytes(base+1, math.Float32bits(0))...) // pos.y
	stream = append(stream, loadXFRegBytes(base+2, math.Float32bits(10))...) // pos.z
	stream = append(stream, loadXFRegBytes(base+3, math.Float32bits(1))...) // color.r
	stream = append(stream, loadXFRegBytes(base+4, math.Float32bits(1))...) // color.g
	stream = append(stream, loadXFRegBytes(base+5, math.Float32bits(1))...) // color.b
	stream = append(stream, loadXFRegBytes(base+6, 1)...)                   // enabled

	stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
	v := Vertex{X: 0, Y: 0, Z: 0, NZ: 1, R: 255, G: 255, B: 255, A: 255}
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V0.R != 255 || tris[0].V0.G != 255 || tris[0].V0.B != 255 {
		t.Fatalf("color = (%d,%d,%d), want (255,255,255): light bound via LOAD_XF_REG facing the normal with white material",
			tris[0].V0.R, tris[0].V0.G, tris[0].V0.B)
	}
}

func TestXFRegistersSelectPositionMatrixIndex(t *testing.T) {
	cp := New()
	translate := []float32{
		1, 0, 0, 20,
		0, 1, 0, 0,
		0, 0, 1, 0,
	}
	var stream []byte
	for i, f := range translate {
		stream = append(stream, loadXFRegBytes(uint16(200+i), math.Float32bits(f))...)
	}
	stream = append(stream, loadXFRegBytes(xfRegPosMatrixIndex, 200)...)

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
	if tris[0].V0.X != 20 {
		t.Fatalf("V0.X = %d, want 20 (matrix selected via xfRegPosMatrixIndex)", tris[0].V0.X)
	}
}
