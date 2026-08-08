package gpu

import "github.com/Duc-inc/espaze/internal/emulation/video"

// Framebuffer is this project's own basic software rasterizer -
// standing in for the Flipper's real hardware rendering pipeline
// (which this project doesn't implement). It fills triangles with
// per-vertex-interpolated (Gouraud) color and depth-tests against a
// Z-buffer; it has no texturing (see tev.go for where a texture
// sample would plug in) and no anti-aliasing.
type Framebuffer struct {
	Width, Height int
	frame         *video.FrameBuffer
	zbuffer       []int32
}

// NewFramebuffer allocates a black frame and a cleared Z-buffer. This
// calls Clear itself: a zero-initialized Z-buffer (Go's default for a
// fresh []int32) reads as "nearest possible", not "far", which would
// make every triangle at Z>=0 fail its very first depth test - every
// caller used to have to remember to call Clear() before the first
// draw to avoid that; now it's not possible to forget.
func NewFramebuffer(width, height int) *Framebuffer {
	f := &Framebuffer{
		Width: width, Height: height,
		frame:   video.NewFrameBuffer(width, height),
		zbuffer: make([]int32, width*height),
	}
	f.Clear()
	return f
}

// Clear resets the color buffer to black and the Z-buffer to "far".
func (f *Framebuffer) Clear() {
	f.frame.Clear(0, 0, 0, 0xFF)
	for i := range f.zbuffer {
		f.zbuffer[i] = 1<<31 - 1
	}
}

// FrameBuffer returns the underlying RGBA frame.
func (f *Framebuffer) FrameBuffer() *video.FrameBuffer { return f.frame }

// edgeFunction is the standard 2D cross-product edge test barycentric
// rasterization is built on.
func edgeFunction(ax, ay, bx, by, px, py int32) int32 {
	return (px-ax)*(by-ay) - (py-ay)*(bx-ax)
}

// DrawTriangle rasterizes one Z-tested triangle. With no texture stage
// bound (every t.Textures entry nil) it's pure Gouraud shading,
// interpolating each vertex's color; each bound stage's texture
// sample is combined into the running color in slot order via its own
// t.TEVOps entry (set from bind.go's bpTevSlot/bpTevOp), the same
// stage-chaining model real hardware's TEV uses - stage 0 combines
// against the lit vertex color, stage 1 against stage 0's output, and
// so on.
func (f *Framebuffer) DrawTriangle(t Triangle) {
	x0, y0, z0 := int32(t.V0.X), int32(t.V0.Y), int32(t.V0.Z)
	x1, y1, z1 := int32(t.V1.X), int32(t.V1.Y), int32(t.V1.Z)
	x2, y2, z2 := int32(t.V2.X), int32(t.V2.Y), int32(t.V2.Z)

	area := edgeFunction(x0, y0, x1, y1, x2, y2)
	if area == 0 {
		return
	}

	minX, maxX := clampRange(min3(x0, x1, x2), max3(x0, x1, x2), int32(f.Width))
	minY, maxY := clampRange(min3(y0, y1, y2), max3(y0, y1, y2), int32(f.Height))

	for py := minY; py < maxY; py++ {
		for px := minX; px < maxX; px++ {
			w0 := edgeFunction(x1, y1, x2, y2, px, py)
			w1 := edgeFunction(x2, y2, x0, y0, px, py)
			w2 := edgeFunction(x0, y0, x1, y1, px, py)
			if !sameSign(w0, w1, w2, area) {
				continue
			}

			b0, b1, b2 := float64(w0)/float64(area), float64(w1)/float64(area), float64(w2)/float64(area)
			depth := int32(b0*float64(z1) + b1*float64(z2) + b2*float64(z0))

			idx := int(py)*f.Width + int(px)
			if depth >= f.zbuffer[idx] {
				continue
			}
			f.zbuffer[idx] = depth

			vertexColor := Color{
				R: byte(b0*float64(t.V1.R) + b1*float64(t.V2.R) + b2*float64(t.V0.R)),
				G: byte(b0*float64(t.V1.G) + b1*float64(t.V2.G) + b2*float64(t.V0.G)),
				B: byte(b0*float64(t.V1.B) + b1*float64(t.V2.B) + b2*float64(t.V0.B)),
				A: 255,
			}

			out := vertexColor
			u := int(b0*float64(t.V1.U) + b1*float64(t.V2.U) + b2*float64(t.V0.U))
			v := int(b0*float64(t.V1.V) + b1*float64(t.V2.V) + b2*float64(t.V0.V))
			for slot, tex := range t.Textures {
				if tex == nil {
					continue
				}
				texel := tex.Sample(u, v)
				out = Combine(t.TEVOps[slot], Color{texel.R, texel.G, texel.B, texel.A}, out)
			}

			f.frame.SetPixel(int(px), int(py), out.R, out.G, out.B, 0xFF)
		}
	}
}

func sameSign(w0, w1, w2, area int32) bool {
	if area > 0 {
		return w0 >= 0 && w1 >= 0 && w2 >= 0
	}
	return w0 <= 0 && w1 <= 0 && w2 <= 0
}

func min3(a, b, c int32) int32 {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

func max3(a, b, c int32) int32 {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}

func clampRange(lo, hi, limit int32) (int32, int32) {
	if lo < 0 {
		lo = 0
	}
	if hi > limit {
		hi = limit
	}
	return lo, hi
}
