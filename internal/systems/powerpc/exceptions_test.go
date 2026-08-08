package powerpc

import "testing"

type stubBus struct{}

func (stubBus) Read8(uint32) byte      { return 0 }
func (stubBus) Read16(uint32) uint16   { return 0 }
func (stubBus) Read32(uint32) uint32   { return 0 }
func (stubBus) Write8(uint32, byte)    {}
func (stubBus) Write16(uint32, uint16) {}
func (stubBus) Write32(uint32, uint32) {}

func TestExternalInterruptJumpsToVectorWhenEESet(t *testing.T) {
	c := New(stubBus{})
	c.SetPC(0x1000)
	c.regs.MSR = MSREE

	c.RaiseExternalInterrupt()

	if c.PC() != ExternalInterruptVector {
		t.Fatalf("PC = %#x, want %#x", c.PC(), ExternalInterruptVector)
	}
	if c.regs.SRR0 != 0x1000 {
		t.Fatalf("SRR0 = %#x, want 0x1000", c.regs.SRR0)
	}
	if c.regs.MSR&MSREE != 0 {
		t.Fatal("MSR EE should be cleared on interrupt entry")
	}
}

func TestExternalInterruptIgnoredWhenEEClear(t *testing.T) {
	c := New(stubBus{})
	c.SetPC(0x1000)

	c.RaiseExternalInterrupt()

	if c.PC() != 0x1000 {
		t.Fatalf("PC = %#x, want unchanged 0x1000", c.PC())
	}
}

func TestRFIRestoresPCAndMSRFromSRR(t *testing.T) {
	c := New(stubBus{})
	c.SetPC(0x1000)
	c.regs.MSR = MSREE
	c.RaiseExternalInterrupt()

	// rfi: primary 19, ext 50
	instr := uint32(19)<<26 | 50<<1
	primaryTable[19](c, instr)

	if c.PC() != 0x1000 {
		t.Fatalf("PC after rfi = %#x, want 0x1000", c.PC())
	}
	if c.regs.MSR&MSREE == 0 {
		t.Fatal("MSR EE should be restored after rfi")
	}
}
