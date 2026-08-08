package xf

import (
	"math"
	"testing"
)

func TestMemorySnapshotRestoreRoundTrip(t *testing.T) {
	mem := NewMemory()
	mem.Write(5, math.Float32bits(1.5))
	regs := NewRegisters()
	regs.PosMatrixIndex = 42
	regs.TexGenCount = 3

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
