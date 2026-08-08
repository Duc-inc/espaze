package gpu

import "github.com/Duc-inc/espaze/internal/emulation/video"

// EncodeXFB converts a rendered RGBA frame into the byte layout real
// GameCube hardware's external framebuffer (XFB) uses: YUY2, a
// standard 4:2:2 packed video format (not GameCube-specific - the same
// format many TV/broadcast video pipelines use), because that's what
// VI's real video encoder reads to drive a display. Two horizontal
// pixels pack into 4 bytes (Y0, U, Y1, V), sharing one averaged U/V
// chroma sample per this format's standard subsampling. An odd
// width's last column reuses its own pixel instead of averaging with
// a neighbor. Luma/chroma use the standard BT.601 RGB-to-YUV
// coefficients, not anything specific to this project.
func EncodeXFB(fb *video.FrameBuffer) []byte {
	out := make([]byte, 0, fb.Width*fb.Height*2)
	for y := 0; y < fb.Height; y++ {
		for x := 0; x < fb.Width; x += 2 {
			y0, u0, v0 := rgbToYUV(fb, x, y)
			x1 := x + 1
			if x1 >= fb.Width {
				x1 = x
			}
			y1, u1, v1 := rgbToYUV(fb, x1, y)
			out = append(out, y0, avgByte(u0, u1), y1, avgByte(v0, v1))
		}
	}
	return out
}

func rgbToYUV(fb *video.FrameBuffer, x, y int) (yy, u, v byte) {
	i := (y*fb.Width + x) * 4
	r, g, b := float64(fb.Pixels[i]), float64(fb.Pixels[i+1]), float64(fb.Pixels[i+2])
	yy = clampByte(0.257*r + 0.504*g + 0.098*b + 16)
	u = clampByte(-0.148*r - 0.291*g + 0.439*b + 128)
	v = clampByte(0.439*r - 0.368*g - 0.071*b + 128)
	return
}

func avgByte(a, b byte) byte { return byte((int(a) + int(b)) / 2) }

func clampByte(v float64) byte {
	switch {
	case v < 0:
		return 0
	case v > 255:
		return 255
	default:
		return byte(v)
	}
}
