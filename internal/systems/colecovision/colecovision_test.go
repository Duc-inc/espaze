package colecovision

import "testing"

// newTestROM builds a minimal ROM: an infinite JP loop at $0000
// (which also appears at $8000, cartridge entry point).
func newTestROM() []byte {
	rom := make([]byte, 0x8000)
	rom[0], rom[1], rom[2] = 0xC3, 0x00, 0x00 // JP $0000
	return rom
}

func TestColecoVisionStepsAFrameWithoutError(t *testing.T) {
	c, ok := New().(*ColecoVision)
	if !ok {
		t.Fatalf("New() did not return *ColecoVision")
	}
	if err := c.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if err := c.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := c.FrameBuffer()
	if fb.Width != 256 || fb.Height != 192 {
		t.Fatalf("unexpected frame size %dx%d", fb.Width, fb.Height)
	}
}

func TestColecoVisionSaveStateRoundTrip(t *testing.T) {
	c, _ := New().(*ColecoVision)
	if err := c.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := c.StepFrame(); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
	}

	data, err := c.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, _ := New().(*ColecoVision)
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM (fresh): %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if fresh.proc.PC() != c.proc.PC() {
		t.Fatalf("restored PC = %#04x, want %#04x", fresh.proc.PC(), c.proc.PC())
	}
}

func TestColecoVisionRejectsEmptyROM(t *testing.T) {
	c, _ := New().(*ColecoVision)
	if err := c.LoadROM(nil); err == nil {
		t.Fatal("expected an error for an empty ROM")
	}
}

func TestColecoVisionDrainAudioReturnsSampleRate(t *testing.T) {
	c, _ := New().(*ColecoVision)
	if err := c.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	if err := c.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	_, rate := c.DrainAudio()
	if rate != 44100 {
		t.Fatalf("sample rate = %d, want 44100", rate)
	}
}
