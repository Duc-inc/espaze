package gpu

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/texture"
)

func TestDrawTriangleFillsInteriorPixel(t *testing.T) {
	f := NewFramebuffer(64, 64)
	f.Clear()

	f.DrawTriangle(Triangle{
		V0: Vertex{X: 10, Y: 10, R: 255, A: 255},
		V1: Vertex{X: 50, Y: 10, G: 255, A: 255},
		V2: Vertex{X: 30, Y: 50, B: 255, A: 255},
	})

	fb := f.FrameBuffer()
	i := (30*fb.Width + 30) * 4
	if fb.Pixels[i] == 0 && fb.Pixels[i+1] == 0 && fb.Pixels[i+2] == 0 {
		t.Fatal("expected a non-black pixel inside the triangle")
	}
}

func TestDrawTriangleLeavesOutsidePixelBlack(t *testing.T) {
	f := NewFramebuffer(64, 64)
	f.Clear()

	f.DrawTriangle(Triangle{
		V0: Vertex{X: 10, Y: 10, R: 255, A: 255},
		V1: Vertex{X: 50, Y: 10, R: 255, A: 255},
		V2: Vertex{X: 30, Y: 50, R: 255, A: 255},
	})

	fb := f.FrameBuffer()
	i := (5*fb.Width + 5) * 4
	if fb.Pixels[i] != 0 || fb.Pixels[i+1] != 0 || fb.Pixels[i+2] != 0 {
		t.Fatal("expected pixel outside the triangle to remain black")
	}
}

func TestZBufferRejectsFartherTriangle(t *testing.T) {
	f := NewFramebuffer(64, 64)
	f.Clear()

	near := Triangle{
		V0: Vertex{X: 0, Y: 0, Z: 0, R: 255, A: 255},
		V1: Vertex{X: 64, Y: 0, Z: 0, R: 255, A: 255},
		V2: Vertex{X: 32, Y: 64, Z: 0, R: 255, A: 255},
	}
	far := Triangle{
		V0: Vertex{X: 0, Y: 0, Z: 100, G: 255, A: 255},
		V1: Vertex{X: 64, Y: 0, Z: 100, G: 255, A: 255},
		V2: Vertex{X: 32, Y: 64, Z: 100, G: 255, A: 255},
	}
	f.DrawTriangle(near)
	f.DrawTriangle(far)

	fb := f.FrameBuffer()
	i := (20*fb.Width + 32) * 4
	if fb.Pixels[i] == 0 || fb.Pixels[i+1] != 0 {
		t.Fatal("expected the nearer (red) triangle to win the Z-test, not the farther green one")
	}
}

func TestDrawTriangleWithTextureModulatesTexelByVertexColor(t *testing.T) {
	f := NewFramebuffer(64, 64)
	f.Clear()

	// A solid blue 1x1 texture, sampled everywhere by a white vertex
	// color - TEVModulate against white should pass the texel through
	// unchanged.
	tex := texture.New(texture.FormatRGBA8, []byte{0, 0, 255, 255}, 1, 1)

	f.DrawTriangle(Triangle{
		V0:       Vertex{X: 0, Y: 0, R: 255, G: 255, B: 255, A: 255},
		V1:       Vertex{X: 64, Y: 0, R: 255, G: 255, B: 255, A: 255},
		V2:       Vertex{X: 32, Y: 64, R: 255, G: 255, B: 255, A: 255},
		Textures: [MaxTexStages]*texture.Texture{tex},
		TEVOps:   [MaxTexStages]TEVOp{TEVModulate},
	})

	fb := f.FrameBuffer()
	i := (20*fb.Width + 32) * 4
	if fb.Pixels[i] != 0 || fb.Pixels[i+1] != 0 || fb.Pixels[i+2] != 255 {
		t.Fatalf("pixel = (%d,%d,%d), want the texture's blue (0,0,255)", fb.Pixels[i], fb.Pixels[i+1], fb.Pixels[i+2])
	}
}

func TestDrawTriangleWithoutTextureIgnoresNilTexture(t *testing.T) {
	f := NewFramebuffer(64, 64)
	f.Clear()

	f.DrawTriangle(Triangle{
		V0: Vertex{X: 0, Y: 0, R: 255, A: 255},
		V1: Vertex{X: 64, Y: 0, R: 255, A: 255},
		V2: Vertex{X: 32, Y: 64, R: 255, A: 255},
	})

	fb := f.FrameBuffer()
	i := (20*fb.Width + 32) * 4
	if fb.Pixels[i] != 255 {
		t.Fatalf("pixel R = %d, want 255 (pure Gouraud, no texture bound)", fb.Pixels[i])
	}
}
