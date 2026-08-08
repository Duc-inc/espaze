package gamecube

import "testing"

func TestLoadAtAndStepExecutesInstruction(t *testing.T) {
	g := New()
	// addi r3, r0, 42 at address 0
	instr := uint32(14)<<26 | 3<<21 | 0<<16 | 42
	g.LoadAt(0, []byte{byte(instr >> 24), byte(instr >> 16), byte(instr >> 8), byte(instr)})

	g.Step()
	if g.proc.PC() != 4 {
		t.Fatalf("PC after one Step = %d, want 4", g.proc.PC())
	}
}

func TestResetClearsState(t *testing.T) {
	g := New()
	g.LoadAt(0, []byte{0, 0, 0, 1})
	g.Step()
	g.Reset()
	if g.proc.PC() != 0 {
		t.Fatalf("PC after Reset = %d, want 0", g.proc.PC())
	}
	if v := g.bus.Read8(0); v != 0 {
		t.Fatalf("MEM1[0] after Reset = %#02x, want 0", v)
	}
}
