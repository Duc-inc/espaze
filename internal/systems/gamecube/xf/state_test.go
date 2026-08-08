package xf

import (
	"math"
	"testing"
)

func TestMemorySnapshotRestoreRoundTrip(t *testing.T) {
	mem := NewMemory()
	mem.Write(5, math.Float32bits(1.5))
	regs := NewRegisters()
	regs.Write(RegMatrixSelection0, 7)

	snap := mem.Snapshot(regs)

	// Mutate after the snapshot to prove Restore reverts it.
	mem.Write(5, math.Float32bits(9.9))

	restoredRegs := mem.Restore(snap)

	if got := mem.ReadFloat32(5); got != 1.5 {
		t.Fatalf("ReadFloat32(5) = %v, want 1.5", got)
	}
	if restoredRegs != regs {
		t.Fatalf("restored registers = %+v, want %+v", restoredRegs, regs)
	}
}

func TestStateLoadSplitsBetweenMemoryAndRegisters(t *testing.T) {
	s := NewState()
	s.Load(0, []uint32{math.Float32bits(3.5)}) // memory (below RegistersStart)
	s.Load(RegMatrixSelection0, []uint32{9})   // control register

	if got := s.Memory.ReadFloat32(0); got != 3.5 {
		t.Fatalf("Memory[0] = %v, want 3.5", got)
	}
	if got := s.Registers.MatrixSelection0.GeometryIndex(); got != 9 {
		t.Fatalf("GeometryIndex = %d, want 9", got)
	}
}

func TestStateLoadIgnoresWordsPastRegistersEnd(t *testing.T) {
	s := NewState()
	s.Load(RegistersEnd, []uint32{0xDEAD}) // must not panic, must not land anywhere meaningful
}
