package powerpc

import "testing"

// crInstr builds a real XL-form CR-logical instruction: BT<<21 |
// BA<<16 | BB<<11 | xo<<1, the same bit positions xForm uses for GPR
// operands, reused here for CR bit indices (0-31, IBM numbering).
func crInstr(bt, ba, bb, xo uint32) uint32 {
	return uint32(19)<<26 | bt<<21 | ba<<16 | bb<<11 | xo<<1
}

func TestCrandCombinesTwoConditionBits(t *testing.T) {
	// crand 10,0,4: CR[10] = CR[0] & CR[4].
	instr := crInstr(10, 0, 4, 257)
	c, _ := newTestCPU([]uint32{instr})
	c.regs.setCRBit(0, true)
	c.regs.setCRBit(4, true)

	c.Step()
	if !c.regs.crBit(10) {
		t.Fatal("expected CR[10] = true (both inputs true)")
	}
}

func TestCrandFalseWhenEitherInputFalse(t *testing.T) {
	instr := crInstr(10, 0, 4, 257)
	c, _ := newTestCPU([]uint32{instr})
	c.regs.setCRBit(0, true)
	c.regs.setCRBit(4, false)

	c.Step()
	if c.regs.crBit(10) {
		t.Fatal("expected CR[10] = false (one input false)")
	}
}

func TestCrorCombinesTwoConditionBits(t *testing.T) {
	// cror 10,0,4: CR[10] = CR[0] | CR[4].
	instr := crInstr(10, 0, 4, 449)
	c, _ := newTestCPU([]uint32{instr})
	c.regs.setCRBit(0, false)
	c.regs.setCRBit(4, true)

	c.Step()
	if !c.regs.crBit(10) {
		t.Fatal("expected CR[10] = true (one input true)")
	}
}

func TestMcrfCopiesWholeField(t *testing.T) {
	// mcrf cr2,cr5: BF=2 (destination), BFA=5 (source).
	instr := uint32(19)<<26 | 2<<23 | 5<<18
	c, _ := newTestCPU([]uint32{instr})
	// Field i occupies Go bits [28-4i, 31-4i] (setCR0's own convention:
	// field 0 = the top 4 bits, shift 28). Field 5 -> shift 8.
	c.regs.CR = 0xA << 8

	c.Step()
	got := (c.regs.CR >> 20) & 0xF // field 2 -> shift 20
	if got != 0xA {
		t.Fatalf("CR field 2 = %#x, want 0xa (copied from field 5)", got)
	}
}
