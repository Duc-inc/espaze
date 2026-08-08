package gpu

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/video"
)

func TestEncodeXFBWhitePixelPair(t *testing.T) {
	fb := video.NewFrameBuffer(2, 1)
	fb.SetPixel(0, 0, 255, 255, 255, 255)
	fb.SetPixel(1, 0, 255, 255, 255, 255)

	out := EncodeXFB(fb)
	if len(out) != 4 {
		t.Fatalf("len = %d, want 4", len(out))
	}
	// White -> Y=235 (16 + 0.859*255), U=V=128 (no chroma offset for gray).
	if out[0] != 235 || out[1] != 128 || out[2] != 235 || out[3] != 128 {
		t.Fatalf("got %v, want [235 128 235 128]", out)
	}
}

func TestEncodeXFBBlackPixelPair(t *testing.T) {
	fb := video.NewFrameBuffer(2, 1)
	// Pixels default to zero (black, alpha 0) from NewFrameBuffer.

	out := EncodeXFB(fb)
	if out[0] != 16 || out[1] != 128 || out[2] != 16 || out[3] != 128 {
		t.Fatalf("got %v, want [16 128 16 128]", out)
	}
}

func TestEncodeXFBOddWidthReusesLastColumn(t *testing.T) {
	fb := video.NewFrameBuffer(1, 1)
	fb.SetPixel(0, 0, 255, 255, 255, 255)

	out := EncodeXFB(fb)
	if len(out) != 4 {
		t.Fatalf("len = %d, want 4 (one packed YUY2 unit even for a 1px-wide frame)", len(out))
	}
	if out[0] != 235 || out[2] != 235 {
		t.Fatalf("got %v, want both Y samples = 235 (duplicated single pixel)", out)
	}
}

func TestEncodeXFBOutputLengthMatchesFrameSize(t *testing.T) {
	fb := video.NewFrameBuffer(640, 480)
	out := EncodeXFB(fb)
	if want := 640 * 480 * 2; len(out) != want {
		t.Fatalf("len = %d, want %d (2 bytes/pixel, YUY2's standard packing)", len(out), want)
	}
}
