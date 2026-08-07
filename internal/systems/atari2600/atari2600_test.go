package atari2600

import "testing"

// newTestROM builds a minimal ROM: an infinite JMP loop at $F000
// (which appears at $1000 too, through the cartridge's mirroring),
// with the reset vector pointing at it.
func newTestROM() []byte {
	rom := make([]byte, 0x1000)
	rom[0x000], rom[0x001], rom[0x002] = 0x4C, 0x00, 0xF0 // JMP $F000
	rom[0xFFC], rom[0xFFD] = 0x00, 0xF0                   // reset vector -> $F000
	return rom
}

func TestAtari2600StepsAFrameWithoutError(t *testing.T) {
	a, ok := New().(*Atari2600)
	if !ok {
		t.Fatalf("New() did not return *Atari2600")
	}
	if err := a.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if err := a.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := a.FrameBuffer()
	if fb.Width != 160 || fb.Height != 192 {
		t.Fatalf("unexpected frame size %dx%d", fb.Width, fb.Height)
	}
}

func TestAtari2600SaveStateRoundTrip(t *testing.T) {
	a, _ := New().(*Atari2600)
	if err := a.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := a.StepFrame(); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
	}

	data, err := a.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, _ := New().(*Atari2600)
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM (fresh): %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if fresh.proc.PC() != a.proc.PC() {
		t.Fatalf("restored PC = %#04x, want %#04x", fresh.proc.PC(), a.proc.PC())
	}
}

func TestAtari2600RejectsEmptyROM(t *testing.T) {
	a, _ := New().(*Atari2600)
	if err := a.LoadROM(nil); err == nil {
		t.Fatal("expected an error for an empty ROM")
	}
}

func TestAtari2600DrainAudioReturnsSampleRate(t *testing.T) {
	a, _ := New().(*Atari2600)
	if err := a.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	if err := a.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	_, rate := a.DrainAudio()
	if rate != 44100 {
		t.Fatalf("sample rate = %d, want 44100", rate)
	}
}
