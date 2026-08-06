package video

// FrameBuffer holds one rendered frame as packed RGBA pixels.
// Every system core renders into a FrameBuffer so the rest of the
// app (engine, app bindings, frontend) never needs to know how a
// specific console produces its picture.
type FrameBuffer struct {
	Width  int
	Height int
	Pixels []byte // RGBA8888, len == Width*Height*4
}

// NewFrameBuffer allocates a black frame of the given size.
func NewFrameBuffer(width, height int) *FrameBuffer {
	return &FrameBuffer{
		Width:  width,
		Height: height,
		Pixels: make([]byte, width*height*4),
	}
}

// Clear resets every pixel to the given RGBA color.
func (f *FrameBuffer) Clear(r, g, b, a byte) {
	for i := 0; i < len(f.Pixels); i += 4 {
		f.Pixels[i] = r
		f.Pixels[i+1] = g
		f.Pixels[i+2] = b
		f.Pixels[i+3] = a
	}
}

// SetPixel writes a single pixel, ignoring out-of-bounds coordinates.
func (f *FrameBuffer) SetPixel(x, y int, r, g, b, a byte) {
	if x < 0 || y < 0 || x >= f.Width || y >= f.Height {
		return
	}
	i := (y*f.Width + x) * 4
	f.Pixels[i] = r
	f.Pixels[i+1] = g
	f.Pixels[i+2] = b
	f.Pixels[i+3] = a
}

// Clone returns an independent copy, safe to hand off across goroutines.
func (f *FrameBuffer) Clone() *FrameBuffer {
	clone := &FrameBuffer{Width: f.Width, Height: f.Height}
	clone.Pixels = make([]byte, len(f.Pixels))
	copy(clone.Pixels, f.Pixels)
	return clone
}
