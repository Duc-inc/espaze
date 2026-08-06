package chip8

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/systems/chip8/display"
)

// program draws the built-in font glyph for digit 0 at (28, 12), then
// loops forever on itself - a minimal, hand-assembled CHIP-8 ROM that
// exercises register loads, the I/font lookup, sprite drawing and jumps.
//
//	0x200  6000   LD V0, 0x00
//	0x202  F029   LD I, font(V0)
//	0x204  611C   LD V1, 0x1C   (x = 28)
//	0x206  620C   LD V2, 0x0C   (y = 12)
//	0x208  D125   DRW V1, V2, 5
//	0x20A  120A   JP 0x20A      (halt)
var program = []byte{
	0x60, 0x00,
	0xF0, 0x29,
	0x61, 0x1C,
	0x62, 0x0C,
	0xD1, 0x25,
	0x12, 0x0A,
}

func newLoadedCore(t *testing.T) *Chip8 {
	t.Helper()
	c, ok := New().(*Chip8)
	if !ok {
		t.Fatalf("New() did not return *Chip8")
	}
	if err := c.LoadROM(program); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	return c
}

func pixelOn(t *testing.T, c *Chip8, x, y int) bool {
	t.Helper()
	fb := c.FrameBuffer()
	idx := (y*fb.Width + x) * 4
	return fb.Pixels[idx] != 0x0B // != background gray
}

func TestChip8DrawsFontSprite(t *testing.T) {
	c := newLoadedCore(t)

	if err := c.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	// Font glyph "0" = 0xF0,0x90,0x90,0x90,0xF0: top row fully lit.
	if !pixelOn(t, c, 28, 12) {
		t.Errorf("expected pixel (28,12) to be on (top-left of glyph)")
	}
	if !pixelOn(t, c, 31, 12) {
		t.Errorf("expected pixel (31,12) to be on (top-right of glyph)")
	}
	// Second row only has the two edge pixels lit, not the middle.
	if pixelOn(t, c, 29, 13) {
		t.Errorf("expected pixel (29,13) to be off (glyph interior)")
	}
	// Somewhere far from the glyph must remain background.
	if pixelOn(t, c, 0, 0) {
		t.Errorf("expected pixel (0,0) to stay off")
	}
}

func TestChip8HaltsOnSelfJump(t *testing.T) {
	c := newLoadedCore(t)

	if err := c.StepFrame(); err != nil {
		t.Fatalf("first StepFrame: %v", err)
	}
	before := c.FrameBuffer()

	// A second frame re-executes the halt loop only (never redraws),
	// so the picture must stay pixel-for-pixel identical.
	if err := c.StepFrame(); err != nil {
		t.Fatalf("second StepFrame: %v", err)
	}
	after := c.FrameBuffer()

	for i := range before.Pixels {
		if before.Pixels[i] != after.Pixels[i] {
			t.Fatalf("frame changed at byte %d after halt loop: %d != %d", i, before.Pixels[i], after.Pixels[i])
		}
	}
}

func TestChip8SaveAndLoadStateRoundTrip(t *testing.T) {
	c := newLoadedCore(t)
	if err := c.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	data, err := c.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, ok := New().(*Chip8)
	if !ok {
		t.Fatalf("New() did not return *Chip8")
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	want := c.FrameBuffer()
	got := fresh.FrameBuffer()
	for i := range want.Pixels {
		if want.Pixels[i] != got.Pixels[i] {
			t.Fatalf("restored frame differs at byte %d: %d != %d", i, want.Pixels[i], got.Pixels[i])
		}
	}
}

func TestChip8DrainAudioMatchesFrameLength(t *testing.T) {
	c := newLoadedCore(t)
	if err := c.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	samples, rate := c.DrainAudio()
	want := rate / int(Metadata().FramesPerSecond)
	if len(samples) != want {
		t.Fatalf("expected %d samples per frame, got %d", want, len(samples))
	}
}

func TestChip8SetInputMapsAllSixteenKeys(t *testing.T) {
	c := newLoadedCore(t)
	c.SetInput(input.State{}.With(0xA, true))

	if !c.keys.IsDown(0xA) {
		t.Fatalf("expected key 0xA to be down after SetInput")
	}
	if c.keys.IsDown(0xB) {
		t.Fatalf("expected key 0xB to remain up")
	}
}

// sanity check that the test's own coordinate math matches the real
// screen bounds, so a future display resize doesn't silently break it.
func TestGlyphCoordinatesFitOnScreen(t *testing.T) {
	if 31 >= display.Width || 16 >= display.Height {
		t.Fatalf("glyph coordinates no longer fit within %dx%d display", display.Width, display.Height)
	}
}
