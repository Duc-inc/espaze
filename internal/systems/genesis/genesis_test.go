package genesis

import "testing"

// newTestROM builds a minimal ROM: a reset vector (SSP, then PC
// pointing at $400) followed by an infinite BRA.S loop at $400.
func newTestROM() []byte {
	rom := make([]byte, 0x1000)
	putLong := func(addr uint32, v uint32) {
		rom[addr] = byte(v >> 24)
		rom[addr+1] = byte(v >> 16)
		rom[addr+2] = byte(v >> 8)
		rom[addr+3] = byte(v)
	}
	putLong(0, 0x00FFFF00) // initial SSP, somewhere in work RAM
	putLong(4, 0x00000400) // initial PC

	rom[0x400], rom[0x401] = 0x60, 0xFE // BRA.S -2 (branch to itself)
	return rom
}

func TestGenesisStepsAFrameWithoutError(t *testing.T) {
	g, ok := New().(*Genesis)
	if !ok {
		t.Fatalf("New() did not return *Genesis")
	}
	if err := g.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	if err := g.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}

	fb := g.FrameBuffer()
	if fb.Width != 320 || fb.Height != 224 {
		t.Fatalf("unexpected frame size %dx%d", fb.Width, fb.Height)
	}
}

func TestGenesisSaveStateRoundTrip(t *testing.T) {
	g, _ := New().(*Genesis)
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

	fresh, _ := New().(*Genesis)
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM (fresh): %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if fresh.cpu.PC() != g.cpu.PC() {
		t.Fatalf("restored PC = %#08x, want %#08x", fresh.cpu.PC(), g.cpu.PC())
	}
}

func TestGenesisRejectsEmptyROM(t *testing.T) {
	g, _ := New().(*Genesis)
	if err := g.LoadROM(nil); err == nil {
		t.Fatal("expected an error for an empty ROM")
	}
}

func TestGenesisDrainAudioReturnsSampleRate(t *testing.T) {
	g, _ := New().(*Genesis)
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
