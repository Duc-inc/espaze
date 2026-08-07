package gamegear

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/input"
)

// newTestROM builds a minimal ROM: an infinite JP loop at $0000.
func newTestROM() []byte {
	rom := make([]byte, 0x8000)
	rom[0], rom[1], rom[2] = 0xC3, 0x00, 0x00 // JP $0000
	return rom
}

func TestGameGearStepsAFrameWithoutError(t *testing.T) {
	g, ok := New().(*GameGear)
	if !ok {
		t.Fatalf("New() did not return *GameGear")
	}
	if err := g.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if err := g.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := g.FrameBuffer()
	if fb.Width != Width || fb.Height != Height {
		t.Fatalf("unexpected frame size %dx%d, want %dx%d", fb.Width, fb.Height, Width, Height)
	}
}

func TestGameGearSaveStateRoundTrip(t *testing.T) {
	g, _ := New().(*GameGear)
	if err := g.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := g.StepFrame(); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
	}

	data, err := g.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, _ := New().(*GameGear)
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM (fresh): %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if fresh.proc.PC() != g.proc.PC() {
		t.Fatalf("restored PC = %#04x, want %#04x", fresh.proc.PC(), g.proc.PC())
	}
}

func TestGameGearRejectsEmptyROM(t *testing.T) {
	g, _ := New().(*GameGear)
	if err := g.LoadROM(nil); err == nil {
		t.Fatal("expected an error for an empty ROM")
	}
}

func TestGameGearStartButtonReadsThroughPortZero(t *testing.T) {
	g, _ := New().(*GameGear)
	if err := g.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if v := g.io.In(0x00); v&0x80 == 0 {
		t.Fatal("Start unpressed should read as bit7 set (active-low, released)")
	}

	g.SetInput(input.State{}.With(Start, true))
	if v := g.io.In(0x00); v&0x80 != 0 {
		t.Fatal("Start pressed should read as bit7 clear (active-low)")
	}
}

func TestGameGearCropsToCenterOfVDPPlane(t *testing.T) {
	g, _ := New().(*GameGear)
	if err := g.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	// Paint a distinct color directly into the underlying VDP's full
	// frame at the crop window's top-left corner, then confirm cropFrame
	// carried it over to (0,0) of the Game Gear's own output.
	g.video.FrameBuffer().SetPixel(offsetX, offsetY, 0x12, 0x34, 0x56, 0xFF)
	g.cropFrame()

	i := 0
	if r, gr, b := g.frame.Pixels[i], g.frame.Pixels[i+1], g.frame.Pixels[i+2]; r != 0x12 || gr != 0x34 || b != 0x56 {
		t.Fatalf("cropped (0,0) = (%#02x,%#02x,%#02x), want (0x12,0x34,0x56)", r, gr, b)
	}
}
