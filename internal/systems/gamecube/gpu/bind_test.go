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

func TestXFRegistersSelectPositionMatrixViaRealMatrixSelection0(t *testing.T) {
	cp := New()
	translate := []float32{
		1, 0, 0, 20,
		0, 1, 0, 0,
		0, 0, 1, 0,
	}
	// Row 1 of the position-matrix block (word address 4, since each
	// row is 4 words) - row 0 stays the default identity matrix New()
	// already wrote, so picking the wrong row would be obvious.
	var stream []byte
	for i, f := range translate {
		stream = append(stream, loadXFRegBytes(uint16(4+i), math.Float32bits(f))...)
	}
	// RegMatrixSelection0 (real XF register 0x1018): GeometryIndex=1 selects row 1.
	stream = append(stream, loadXFRegBytes(xf.RegMatrixSelection0, 1)...)

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
		t.Fatalf("V0.X = %d, want 20 (matrix selected via the real RegMatrixSelection0 register)", tris[0].V0.X)
	}
}

func TestXFAmbientColorRegisterWriteDrivesRealLighting(t *testing.T) {
	cp := New()
	cp.SetLight(0, xf.Light{}) // no lights: isolate ambient's own contribution

	stream := loadXFRegBytes(xf.RegAmbientColor0, 0x80000000) // R=128, G=B=0
	stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
	v := Vertex{X: 0, Y: 0, Z: 0, R: 255, G: 255, B: 255, A: 255}
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	stream = append(stream, vertexBytesOf(v)...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].V0.R != 128 || tris[0].V0.G != 0 || tris[0].V0.B != 0 {
		t.Fatalf("color = (%d,%d,%d), want (128,0,0): a real LOAD_XF_REG write to RegAmbientColor0 should drive lighting, not just SetAmbient",
			tris[0].V0.R, tris[0].V0.G, tris[0].V0.B)
	}
}

func TestXFLightMemoryWriteDrivesRealLighting(t *testing.T) {
	cp := New()
	// Identity normal matrix at the default GeometryIndex(0) address.
	cp.xfState.Memory.WriteNormalMatrix(xf.NormalMatricesStart, xf.NormalMatrix{
		{1, 0, 0}, {0, 1, 0}, {0, 0, 1},
	})
	cp.SetAmbient(xf.Ambient{}) // isolate the light's own contribution

	lightBase := uint16(xf.LightsStart) // light 0
	var stream []byte
	stream = append(stream, loadXFRegBytes(lightBase+xf.LightColorOffset, 0xFFFFFF00)...) // R=G=B=255
	stream = append(stream, loadXFRegBytes(lightBase+xf.LightPositionOffset, math.Float32bits(0))...)
	stream = append(stream, loadXFRegBytes(lightBase+xf.LightPositionOffset+1, math.Float32bits(0))...)
	stream = append(stream, loadXFRegBytes(lightBase+xf.LightPositionOffset+2, math.Float32bits(10))...)

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
		t.Fatalf("color = (%d,%d,%d), want (255,255,255): a real LOAD_XF_REG write into XF light memory should drive lighting, not just SetLight",
			tris[0].V0.R, tris[0].V0.G, tris[0].V0.B)
	}
}
