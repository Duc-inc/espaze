package nes

import "testing"

// newTestROM builds a minimal 32KB NROM (mapper 0) image: an infinite
// JMP loop at $8000, with the reset vector pointing at it.
func newTestROM() []byte {
	header := []byte{'N', 'E', 'S', 0x1A, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	prg := make([]byte, 32*1024)
	prg[0], prg[1], prg[2] = 0x4C, 0x00, 0x80 // JMP $8000
	prg[0x7FFC], prg[0x7FFD] = 0x00, 0x80     // reset vector -> $8000
	chr := make([]byte, 8*1024)

	rom := append(header, prg...)
	rom = append(rom, chr...)
	return rom
}

func TestNESStepsAFrameWithoutError(t *testing.T) {
	n, ok := New().(*NES)
	if !ok {
		t.Fatalf("New() did not return *NES")
	}
	if err := n.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if err := n.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := n.FrameBuffer()
	if fb.Width != 256 || fb.Height != 240 {
		t.Fatalf("unexpected frame size %dx%d", fb.Width, fb.Height)
	}
}

func TestNESSaveStateRoundTrip(t *testing.T) {
	n, _ := New().(*NES)
	if err := n.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := n.StepFrame(); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
	}

	data, err := n.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, _ := New().(*NES)
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM (fresh): %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if fresh.proc.PC() != n.proc.PC() {
		t.Fatalf("restored PC = %#04x, want %#04x", fresh.proc.PC(), n.proc.PC())
	}
}

func TestNESRejectsInvalidHeader(t *testing.T) {
	n, _ := New().(*NES)
	if err := n.LoadROM([]byte("not a rom")); err == nil {
		t.Fatal("expected an error for a non-iNES file")
	}
}
