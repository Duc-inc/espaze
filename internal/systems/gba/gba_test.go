package gba

import "testing"

// newTestROM builds a minimal ROM: an infinite ARM branch-to-self at
// the entry point.
func newTestROM() []byte {
	rom := make([]byte, 0x1000)
	// B $+0 (branch to self): cond=AL, 1010, offset=0xFFFFFE (-2 words)
	rom[0], rom[1], rom[2], rom[3] = 0xFE, 0xFF, 0xFF, 0xEA
	return rom
}

func TestGBAStepsAFrameWithoutError(t *testing.T) {
	g, ok := New().(*GBA)
	if !ok {
		t.Fatalf("New() did not return *GBA")
	}
	if err := g.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if err := g.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := g.FrameBuffer()
	if fb.Width != 240 || fb.Height != 160 {
		t.Fatalf("unexpected frame size %dx%d", fb.Width, fb.Height)
	}
}

func TestGBASaveStateRoundTrip(t *testing.T) {
	g, _ := New().(*GBA)
	if err := g.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := g.StepFrame(); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
	}

	data, err := g.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, _ := New().(*GBA)
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM (fresh): %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if fresh.proc.PC() != g.proc.PC() {
		t.Fatalf("restored PC = %#08x, want %#08x", fresh.proc.PC(), g.proc.PC())
	}
}

func TestGBARejectsEmptyROM(t *testing.T) {
	g, _ := New().(*GBA)
	if err := g.LoadROM(nil); err == nil {
		t.Fatal("expected an error for an empty ROM")
	}
}

func TestGBADrainAudioReturnsSampleRate(t *testing.T) {
	g, _ := New().(*GBA)
	if err := g.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	if err := g.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	_, rate := g.DrainAudio()
	if rate != 44100 {
		t.Fatalf("sample rate = %d, want 44100", rate)
	}
}
