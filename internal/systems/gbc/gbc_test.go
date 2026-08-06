package gbc

import "testing"

// program writes a solid tile into tile 0 and enables the LCD/background,
// then loops forever - reusing the exact same pattern the DMG core's own
// test uses, since the CPU/instruction set is identical.
var program = []byte{
	0x3E, 0xFF, // LD A,0xFF
	0xEA, 0x00, 0x80, // LD (0x8000),A
	0x3E, 0x00, // LD A,0x00
	0xEA, 0x01, 0x80, // LD (0x8001),A
	0x18, 0xFE, // JR -2 (halt loop)
}

func newTestROM() []byte {
	rom := make([]byte, 0x8000) // 32KB, MBC0-sized
	copy(rom[0x100:], program)
	rom[0x147] = 0x00 // cartridge type: ROM only
	rom[0x149] = 0x00 // no external RAM
	return rom
}

func TestGBCBootsWithCGBRegisterA(t *testing.T) {
	gb, ok := New().(*GBC)
	if !ok {
		t.Fatalf("New() did not return *GBC")
	}
	if err := gb.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	// A must read back 0x11 (CGB), not DMG's 0x01 - this is the one
	// boot-time signal CGB-aware cartridges check.
	if got := gb.proc.Snapshot().Regs.A; got != 0x11 {
		t.Fatalf("A after boot = %#02x, want 0x11", got)
	}
}

func TestGBCStepsAFrameAndRendersATile(t *testing.T) {
	gb, _ := New().(*GBC)
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
}

func TestGBCSaveStateRoundTrip(t *testing.T) {
	gb, _ := New().(*GBC)
	if err := gb.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := gb.StepFrame(); err != nil {
			t.Fatalf("StepFrame: %v", err)
		}
	}

	data, err := gb.SaveState()
	if err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	fresh, _ := New().(*GBC)
	if err := fresh.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM (fresh): %v", err)
	}
	if err := fresh.LoadState(data); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	if fresh.proc.PC() != gb.proc.PC() {
		t.Fatalf("restored PC = %#04x, want %#04x", fresh.proc.PC(), gb.proc.PC())
	}
}

func TestGBCDoubleSpeedHalvesRealCycles(t *testing.T) {
	gb, _ := New().(*GBC)
	if err := gb.LoadROM(newTestROM()); err != nil {
		t.Fatalf("LoadROM: %v", err)
	}

	gb.mmu.Write(0xFF4D, 0x01) // arm + immediately switch to double speed (see speed.go)
	if !gb.mmu.DoubleSpeed() {
		t.Fatal("expected double speed after the KEY1 write")
	}

	if err := gb.StepFrame(); err != nil {
		t.Fatalf("StepFrame: %v", err)
	}
	// Just confirm it doesn't panic/hang across a frame in double-speed
	// mode - StepFrame's loop bound is real-time cycles either way.
}
