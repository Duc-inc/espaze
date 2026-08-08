package gpu

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/texture"
)

func TestBPRegistersBindTwoTextureStagesChained(t *testing.T) {
	cp := New()
	mem := &fakeMemory{data: make([]byte, 0x300)}
	copy(mem.data[0x100:], []byte{255, 0, 0, 255}) // stage 0: red
	copy(mem.data[0x200:], []byte{0, 255, 0, 255}) // stage 1: green
	cp.SetMemoryReader(mem)

	var stream []byte
	// Stage 0 (default active slot): red, Modulate.
	stream = append(stream, loadBPRegBytes(bpTexAddr, 0x100)...)
	stream = append(stream, loadBPRegBytes(bpTexFormat, uint32(texture.FormatRGBA8))...)
	stream = append(stream, loadBPRegBytes(bpTexWidth, 1)...)
	stream = append(stream, loadBPRegBytes(bpTexHeight, 1)...)
	stream = append(stream, loadBPRegBytes(bpTevOp, uint32(TEVModulate))...)
	// Stage 1: green, Add.
	stream = append(stream, loadBPRegBytes(bpTexSlot, 1)...)
	stream = append(stream, loadBPRegBytes(bpTexAddr, 0x200)...)
	stream = append(stream, loadBPRegBytes(bpTexFormat, uint32(texture.FormatRGBA8))...)
	stream = append(stream, loadBPRegBytes(bpTexWidth, 1)...)
	stream = append(stream, loadBPRegBytes(bpTexHeight, 1)...)
	stream = append(stream, loadBPRegBytes(bpTevOp, uint32(TEVAdd))...)

	stream = append(stream, cmdDrawTriangles, 0x00, 0x03)
	v0 := Vertex{X: 0, Y: 0, R: 255, G: 255, B: 255, A: 255}
	v1 := Vertex{X: 64, Y: 0, R: 255, G: 255, B: 255, A: 255}
	v2 := Vertex{X: 32, Y: 64, R: 255, G: 255, B: 255, A: 255}
	stream = append(stream, vertexBytesOf(v0)...)
	stream = append(stream, vertexBytesOf(v1)...)
	stream = append(stream, vertexBytesOf(v2)...)
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1", len(tris))
	}
	if tris[0].Textures[0] == nil || tris[0].Textures[1] == nil {
		t.Fatalf("expected both stage 0 and stage 1 bound, got %v / %v", tris[0].Textures[0], tris[0].Textures[1])
	}

	f := NewFramebuffer(64, 64)
	f.Clear()
	f.DrawTriangle(tris[0])
	fb := f.FrameBuffer()
	i := (20*fb.Width + 32) * 4
	// Stage 0: white vertex modulated by red texel = red. Stage 1:
	// red + green (Add) = yellow.
	if fb.Pixels[i] != 255 || fb.Pixels[i+1] != 255 || fb.Pixels[i+2] != 0 {
		t.Fatalf("pixel = (%d,%d,%d), want yellow (255,255,0) from chained stages", fb.Pixels[i], fb.Pixels[i+1], fb.Pixels[i+2])
	}
}

func TestBPTexSlotOutOfRangeWrapsRatherThanPanicking(t *testing.T) {
	cp := New()
	// 255 % MaxTexStages must stay in range - must not panic.
	cp.Execute(loadBPRegBytes(bpTexSlot, 255))
	cp.Execute(loadBPRegBytes(bpTexAddr, 0x100))
}
