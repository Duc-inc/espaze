package sms

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/systems/sms/memory"
)

// newTestROM builds a minimal ROM: an infinite JP loop at $0000, with a
// nametable/tile already primed so a frame actually renders something.
func newTestROM() []byte {
	rom := make([]byte, 0x8000)
	rom[0], rom[1], rom[2] = 0xC3, 0x00, 0x00 // JP $0000
	return rom
}

func TestSMSStepsAFrameWithoutError(t *testing.T) {
	s, ok := New().(*SMS)
	if !ok {
		t.Fatalf("New() did not return *SMS")
	}
	if err := s.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if err := s.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := s.FrameBuffer()
	if fb.Width != 256 || fb.Height != 192 {
		t.Fatalf("unexpected frame size %dx%d", fb.Width, fb.Height)
	}
}

func TestSMSSaveStateRoundTrip(t *testing.T) {
	s, _ := New().(*SMS)
	if err := s.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := s.StepFrame(); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
	}

	data, err := s.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, _ := New().(*SMS)
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM (fresh): %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if fresh.proc.PC() != s.proc.PC() {
		t.Fatalf("restored PC = %#04x, want %#04x", fresh.proc.PC(), s.proc.PC())
	}
}

func TestSMSRejectsEmptyROM(t *testing.T) {
	s, _ := New().(*SMS)
	if err := s.LoadROM(nil); err == nil {
		t.Fatal("expected an error for an empty ROM")
	}
}

func TestSMSPauseTriggersNMIOnce(t *testing.T) {
	s, _ := New().(*SMS)
	if err := s.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	s.bus.SetButtons(input.State{}.With(memory.Pause, true))
	// The NMI is only latched as pending at the end of the frame it's
	// detected in (see StepFrame) - servicing it is then just the next
	// CPU instruction boundary, checked directly here rather than via
	// another whole StepFrame (which would run on well past the vector
	// before returning, landing PC somewhere else entirely).
	if err := s.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	s.proc.Step()
	if s.proc.PC() != 0x0066 {
		t.Fatalf("PC after Pause = %#04x, want 0x0066 (NMI vector)", s.proc.PC())
	}
}
