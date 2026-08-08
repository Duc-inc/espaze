package gpu

import "testing"

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
