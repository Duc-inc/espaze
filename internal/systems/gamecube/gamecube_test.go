package gamecube

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/ai"
	"github.com/Duc-inc/espaze/internal/systems/powerpc"
)

func TestLoadAtAndStepExecutesInstruction(t *testing.T) {
	g := New(nil)
	// addi r3, r0, 42 at address 0
	instr := uint32(14)<<26 | 3<<21 | 0<<16 | 42
	g.LoadAt(0, []byte{byte(instr >> 24), byte(instr >> 16), byte(instr >> 8), byte(instr)})

	g.Step()
	if g.proc.PC() != 4 {
		t.Fatalf("PC after one Step = %d, want 4", g.proc.PC())
	}
}

func TestPeripheralRegistersAreReachableFromRealPowerPCCode(t *testing.T) {
	g := New(nil)
	// addi r3,r0,0x0001 ; addi r4,r0,(AI.Base low bits offset 0) via lis/ori
	// simplest: use SetGPR to seed the address, then stw through it.
	g.proc.SetGPR(4, ai.Base)               // r4 = AI register block base (AICR)
	g.proc.SetGPR(3, 1)                     // r3 = PSTAT bit
	instr := uint32(36)<<26 | 3<<21 | 4<<16 // stw r3,0(r4)
	g.LoadAt(0, []byte{byte(instr >> 24), byte(instr >> 16), byte(instr >> 8), byte(instr)})

	g.Step()

	if !g.AI.Playing() {
		t.Fatal("expected a real stw through the CPU to reach AI's AICR register")
	}
}

func TestVIVBlankRaisesRealExternalInterrupt(t *testing.T) {
	g := New(nil)
	g.proc.SetMSR(powerpc.MSREE)
	g.VI.Write32(0x30, 1<<28|1) // DI0: enabled, target line 1 (vpos after one Step)

	g.Step()

	if g.proc.PC() != powerpc.ExternalInterruptVector {
		t.Fatalf("PC after VBlank = %#x, want external interrupt vector %#x", g.proc.PC(), powerpc.ExternalInterruptVector)
	}
}

func TestResetClearsState(t *testing.T) {
	g := New(nil)
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
