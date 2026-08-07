package pcengine

import "testing"

// newTestROM builds a minimal ROM: an infinite branch-to-self loop at
// $0000, with the reset vector (in the hardware I/O page, physical
// page 0xFF) pointing at it.
func newTestROM() []byte {
	rom := make([]byte, 0x1000)
	rom[0], rom[1] = 0x80, 0xFE // BRA -2 (branch to itself)
	return rom
}

func TestPCEngineStepsAFrameWithoutError(t *testing.T) {
	p, ok := New().(*PCEngine)
	if !ok {
		t.Fatalf("New() did not return *PCEngine")
	}
	if err := p.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if err := p.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := p.FrameBuffer()
	if fb.Width != 256 || fb.Height != 224 {
		t.Fatalf("unexpected frame size %dx%d", fb.Width, fb.Height)
	}
}

func TestPCEngineSaveStateRoundTrip(t *testing.T) {
	p, _ := New().(*PCEngine)
	if err := p.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := p.StepFrame(); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
	}

	data, err := p.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, _ := New().(*PCEngine)
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM (fresh): %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if fresh.proc.PC() != p.proc.PC() {
		t.Fatalf("restored PC = %#04x, want %#04x", fresh.proc.PC(), p.proc.PC())
	}
}

func TestPCEngineRejectsEmptyROM(t *testing.T) {
	p, _ := New().(*PCEngine)
	if err := p.LoadROM(nil); err == nil {
		t.Fatal("expected an error for an empty ROM")
	}
}

func TestPCEngineDrainAudioReturnsSampleRate(t *testing.T) {
	p, _ := New().(*PCEngine)
	if err := p.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	if err := p.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	_, rate := p.DrainAudio()
	if rate != 44100 {
		t.Fatalf("sample rate = %d, want 44100", rate)
	}
}
