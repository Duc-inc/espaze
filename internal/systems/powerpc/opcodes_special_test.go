package powerpc

import "testing"

func mftbInstr(rD, tbr uint32) uint32 {
	return uint32(31)<<26 | rD<<21 | (tbr&0x1F)<<16 | ((tbr>>5)&0x1F)<<11 | 371<<1
}

func TestMftbReadsTimeBaseLowAndHigh(t *testing.T) {
	mftbl := mftbInstr(3, 268)
	mftbu := mftbInstr(4, 269)
	c, _ := newTestCPU([]uint32{mftbl, mftbu})
	c.regs.TB = 0x1_0000_0005 // high=1, low=5

	c.Step() // consumes TB++ too, but we set TB directly right before, so it's exact pre-Step value +1 after
	if c.regs.GPR[3] != 6 {
		t.Fatalf("GPR3 (TBL) = %d, want 6 (TB incremented once by this Step before mftbl read it)", c.regs.GPR[3])
	}
	c.Step()
	if c.regs.GPR[4] != 1 {
		t.Fatalf("GPR4 (TBU) = %d, want 1", c.regs.GPR[4])
	}
}

func TestStepAdvancesTimeBase(t *testing.T) {
	c, _ := newTestCPU([]uint32{0, 0, 0})
	if c.regs.TB != 0 {
		t.Fatalf("TB = %d, want 0 before any Step", c.regs.TB)
	}
	c.Step()
	c.Step()
	c.Step()
	if c.regs.TB != 3 {
		t.Fatalf("TB = %d, want 3 after 3 Steps", c.regs.TB)
	}
}
