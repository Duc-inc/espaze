package snes

import "testing"

// newTestROM builds a minimal ROM: an infinite branch-to-self loop at
// $8000 (bank 0), with the reset vector pointing at it.
func newTestROM() []byte {
	rom := make([]byte, 0x8000)
	rom[0x7FFC], rom[0x7FFD] = 0x00, 0x80 // reset vector -> $8000
	rom[0], rom[1] = 0x80, 0xFE           // BRA -2 (branch to itself)
	return rom
}

func TestSNESStepsAFrameWithoutError(t *testing.T) {
	s, ok := New().(*SNES)
	if !ok {
		t.Fatalf("New() did not return *SNES")
	}
	if err := s.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if err := s.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := s.FrameBuffer()
	if fb.Width != 256 || fb.Height != 224 {
		t.Fatalf("unexpected frame size %dx%d", fb.Width, fb.Height)
	}
}

func TestSNESSaveStateRoundTrip(t *testing.T) {
	s, _ := New().(*SNES)
	if err := s.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := s.StepFrame(); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
	}

	data, err := s.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, _ := New().(*SNES)
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM (fresh): %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if fresh.proc.PC() != s.proc.PC() {
		t.Fatalf("restored PC = %#06x, want %#06x", fresh.proc.PC(), s.proc.PC())
	}
}

func TestSNESRejectsEmptyROM(t *testing.T) {
	s, _ := New().(*SNES)
	if err := s.LoadROM(nil); err == nil {
		t.Fatal("expected an error for an empty ROM")
	}
}

func TestSNESDrainAudioReturnsSampleRate(t *testing.T) {
	s, _ := New().(*SNES)
	if err := s.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	if err := s.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	_, rate := s.DrainAudio()
	if rate != 44100 {
		t.Fatalf("sample rate = %d, want 44100", rate)
	}
}
