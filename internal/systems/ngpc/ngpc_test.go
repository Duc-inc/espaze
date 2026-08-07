package ngpc

import "testing"

// newTestROM builds a minimal ROM: an infinite JR loop at $0000 (JR
// #-2, opcode 0xA0 in this project's own encoding, branches to itself).
func newTestROM() []byte {
	rom := make([]byte, 0x1000)
	rom[0], rom[1] = 0xA0, 0xFE
	return rom
}

func TestNGPCStepsAFrameWithoutError(t *testing.T) {
	n, ok := New().(*NGPC)
	if !ok {
		t.Fatalf("New() did not return *NGPC")
	}
	if err := n.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if err := n.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := n.FrameBuffer()
	if fb.Width != 160 || fb.Height != 152 {
		t.Fatalf("unexpected frame size %dx%d", fb.Width, fb.Height)
	}
}

func TestNGPCSaveStateRoundTrip(t *testing.T) {
	n, _ := New().(*NGPC)
	if err := n.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := n.StepFrame(); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
	}

	data, err := n.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, _ := New().(*NGPC)
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM (fresh): %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if fresh.proc.PC() != n.proc.PC() {
		t.Fatalf("restored PC = %#06x, want %#06x", fresh.proc.PC(), n.proc.PC())
	}
}

func TestNGPCRejectsEmptyROM(t *testing.T) {
	n, _ := New().(*NGPC)
	if err := n.LoadROM(nil); err == nil {
		t.Fatal("expected an error for an empty ROM")
	}
}

func TestNGPCDrainAudioReturnsSampleRate(t *testing.T) {
	n, _ := New().(*NGPC)
	if err := n.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	if err := n.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	_, rate := n.DrainAudio()
	if rate != 44100 {
		t.Fatalf("sample rate = %d, want 44100", rate)
	}
}
