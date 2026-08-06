package gameboy

import "testing"

// program writes an 8-pixel-wide row of "color index 1" pixels into tile
// 0's first row, then loops forever. Reset() already leaves LCDC/BGP at
// their real post-boot values (LCD+background on, tile map defaulting to
// tile 0 everywhere since VRAM starts zeroed), so this is enough to make
// every 8th scanline (the top row of each background tile) solid black
// and the rest white - a simple, fully predictable picture to assert on.
//
//	0x100  3E FF      LD A,0xFF
//	0x102  EA 00 80    LD (0x8000),A
//	0x105  3E 00        LD A,0x00
//	0x107  EA 01 80      LD (0x8001),A
//	0x10A  18 FE          JR -2 (halt loop)
var program = []byte{
	0x3E, 0xFF,
	0xEA, 0x00, 0x80,
	0x3E, 0x00,
	0xEA, 0x01, 0x80,
	0x18, 0xFE,
}

func newTestROM() []byte {
	rom := make([]byte, 0x8000) // 32KB, MBC0-sized
	copy(rom[0x100:], program)
	rom[0x147] = 0x00 // cartridge type: ROM only
	rom[0x149] = 0x00 // no external RAM
	return rom
}

func TestGameBoyRendersWrittenTile(t *testing.T) {
	gb, ok := New().(*GameBoy)
	if !ok {
		t.Fatalf("New() did not return *GameBoy")
	}
	if err := gb.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if err := gb.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := gb.FrameBuffer()
	if fb.Width != 160 || fb.Height != 144 {
		t.Fatalf("unexpected frame size %dx%d", fb.Width, fb.Height)
	}

	isDark := func(x, y int) bool {
		idx := (y*fb.Width + x) * 4
		return fb.Pixels[idx] < 0x80 // shade 3 (black) vs shade 0 (white)
	}

	if !isDark(0, 0) {
		t.Errorf("expected (0,0) dark: top row of every tile was written to color index 1")
	}
	if !isDark(159, 0) {
		t.Errorf("expected (159,0) dark: the whole top row of the screen should be tile row 0")
	}
	if isDark(0, 1) {
		t.Errorf("expected (0,1) light: tile row 1 was never written, still color index 0")
	}
	if !isDark(0, 8) {
		t.Errorf("expected (0,8) dark: that's row 0 of the *next* tile down")
	}
}

func TestGameBoyMetadataMatchesScreenSize(t *testing.T) {
	meta := Metadata()
	if meta.ScreenWidth != 160 || meta.ScreenHeight != 144 {
		t.Fatalf("metadata screen size %dx%d, want 160x144", meta.ScreenWidth, meta.ScreenHeight)
	}
	if meta.ID != "gameboy" {
		t.Fatalf("metadata ID = %q, want \"gameboy\"", meta.ID)
	}
}

func TestGameBoySaveAndLoadStateRoundTrip(t *testing.T) {
	gb, ok := New().(*GameBoy)
	if !ok {
		t.Fatalf("New() did not return *GameBoy")
	}
	if err := gb.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	if err := gb.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	data, err := gb.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, ok := New().(*GameBoy)
	if !ok {
		t.Fatalf("New() did not return *GameBoy")
	}
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("fresh LoadROM: %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	want, got := gb.FrameBuffer(), fresh.FrameBuffer()
	for i := range want.Pixels {
		if want.Pixels[i] != got.Pixels[i] {
			t.Fatalf("restored frame differs at byte %d", i)
		}
	}
}
