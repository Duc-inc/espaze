package schip

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/video"
	"github.com/Duc-inc/espaze/internal/systems/schip/display"
)

func newLoadedCore(t *testing.T, program []byte) *Schip {
	t.Helper()
	c, ok := New().(*Schip)
	if !ok {
		t.Fatalf("New() did not return *Schip")
	}
	if err := c.LoadROM(program); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	return c
}

func pixelOn(fb *video.FrameBuffer, x, y int) bool {
	idx := (y*fb.Width + x) * 4
	return fb.Pixels[idx] != 0x0B
}

// program: 00FF (extended mode), draw a 16x16 sprite (DXY0) whose data is
// appended right after the halt loop, at (0,0). Only the top row is lit.
//
//	0x200  00FF   high-res on
//	0x202  6100   V1 = 0            (x)
//	0x204  6200   V2 = 0            (y)
//	0x206  A20C   I = 0x20C         (sprite data)
//	0x208  D120   DRW V1, V2, 0     (16x16 sprite)
//	0x20A  120A   JP 0x20A          (halt)
//	0x20C  ..     32 bytes of sprite data
var wideDrawProgram = buildWideDrawProgram()

func buildWideDrawProgram() []byte {
	prog := []byte{
		0x00, 0xFF,
		0x61, 0x00,
		0x62, 0x00,
		0xA2, 0x0C,
		0xD1, 0x20,
		0x12, 0x0A,
	}
	sprite := make([]byte, 32)
	sprite[0], sprite[1] = 0xFF, 0xFF // top row fully lit (16 pixels)
	return append(prog, sprite...)
}

func TestSchipExtendedModeAndWideSprite(t *testing.T) {
	c := newLoadedCore(t, wideDrawProgram)

	if err := c.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := c.FrameBuffer()
	if fb.Width != display.HighWidth || fb.Height != display.HighHeight {
		t.Fatalf("expected %dx%d frame after 00FF, got %dx%d", display.HighWidth, display.HighHeight, fb.Width, fb.Height)
	}

	if !pixelOn(fb, 0, 0) {
		t.Errorf("expected pixel (0,0) on")
	}
	if !pixelOn(fb, 15, 0) {
		t.Errorf("expected pixel (15,0) on (16-wide sprite)")
	}
	if pixelOn(fb, 0, 1) {
		t.Errorf("expected pixel (0,1) off (only top row was lit)")
	}
}

// program: store V0=5,V1=10 into RPL flags, clear both, reload from RPL,
// then draw the font digit "0" at (V0,V1) - only correct if the RPL
// round trip actually restored the original values.
var rplProgram = []byte{
	0x60, 0x05, // LD V0, 5
	0x61, 0x0A, // LD V1, 10
	0xF1, 0x75, // LD R, V1   (store V0..V1)
	0x60, 0x00, // LD V0, 0
	0x61, 0x00, // LD V1, 0
	0xF1, 0x85, // LD V1, R   (load V0..V1)
	0x62, 0x00, // LD V2, 0
	0xF2, 0x29, // LD I, font(V2)
	0xD0, 0x15, // DRW V0, V1, 5
	0x12, 0x12, // JP 0x212 (halt)
}

func TestSchipRplFlagsRoundTrip(t *testing.T) {
	c := newLoadedCore(t, rplProgram)

	if err := c.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := c.FrameBuffer()

	if !pixelOn(fb, 5, 10) {
		t.Errorf("expected pixel (5,10) on: RPL restore should have set V0=5,V1=10")
	}
	if pixelOn(fb, 9, 10) {
		t.Errorf("expected pixel (9,10) off (outside the glyph's lit columns)")
	}
}

// program: draw a single pixel at (5,5), then scroll the screen down 2
// rows; the pixel should end up at (5,7) and no longer be at (5,5).
var scrollProgram = []byte{
	0x60, 0x05, // LD V0, 5
	0x61, 0x05, // LD V1, 5
	0xA2, 0x0C, // LD I, 0x20C (sprite data)
	0xD0, 0x11, // DRW V0, V1, 1
	0x00, 0xC2, // scroll down 2
	0x12, 0x0A, // JP 0x20A (halt)
	0x80, // sprite data: leftmost pixel only
}

func TestSchipScrollDown(t *testing.T) {
	c := newLoadedCore(t, scrollProgram)

	if err := c.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := c.FrameBuffer()

	if pixelOn(fb, 5, 5) {
		t.Errorf("expected pixel (5,5) off after scrolling down")
	}
	if !pixelOn(fb, 5, 7) {
		t.Errorf("expected pixel (5,7) on after scrolling the drawn pixel down by 2")
	}
}

func TestSchipSaveAndLoadStatePreservesResolution(t *testing.T) {
	c := newLoadedCore(t, wideDrawProgram)
	if err := c.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	data, err := c.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, ok := New().(*Schip)
	if !ok {
		t.Fatalf("New() did not return *Schip")
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	want := c.FrameBuffer()
	got := fresh.FrameBuffer()
	if want.Width != got.Width || want.Height != got.Height {
		t.Fatalf("resolution not preserved: want %dx%d, got %dx%d", want.Width, want.Height, got.Width, got.Height)
	}
	for i := range want.Pixels {
		if want.Pixels[i] != got.Pixels[i] {
			t.Fatalf("restored frame differs at byte %d", i)
		}
	}
}
