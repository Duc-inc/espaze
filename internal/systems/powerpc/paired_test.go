package powerpc

import "testing"

func TestPsAddOperatesOnBothSlots(t *testing.T) {
	// ps_add f3,f1,f2 (ext4A=21)
	instr := uint32(4)<<26 | 3<<21 | 1<<16 | 2<<11 | 21<<1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.FPR[1], c.regs.PS1[1] = 1.5, 10
	c.regs.FPR[2], c.regs.PS1[2] = 2.25, 20

	c.Step()
	if c.regs.FPR[3] != 3.75 {
		t.Fatalf("FPR3 = %v, want 3.75", c.regs.FPR[3])
	}
	if c.regs.PS1[3] != 30 {
		t.Fatalf("PS1[3] = %v, want 30", c.regs.PS1[3])
	}
}

func TestPsMulUsesFRCOnBothSlots(t *testing.T) {
	// ps_mul f3,f1,f2 (ext4A=25): frA=1, frC=2
	instr := uint32(4)<<26 | 3<<21 | 1<<16 | 0<<11 | 2<<6 | 25<<1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.FPR[1], c.regs.PS1[1] = 4, 2
	c.regs.FPR[2], c.regs.PS1[2] = 2.5, 3

	c.Step()
	if c.regs.FPR[3] != 10 {
		t.Fatalf("FPR3 = %v, want 10", c.regs.FPR[3])
	}
	if c.regs.PS1[3] != 6 {
		t.Fatalf("PS1[3] = %v, want 6", c.regs.PS1[3])
	}
}

func TestPsNegAndAbsOperateOnBothSlots(t *testing.T) {
	psNeg := uint32(4)<<26 | 1<<21 | 0<<16 | 2<<11 | 40<<1
	psAbs := uint32(4)<<26 | 3<<21 | 0<<16 | 1<<11 | 264<<1
	c, _ := newTestCPU([]uint32{psNeg, psAbs})
	c.regs.FPR[2], c.regs.PS1[2] = 5, 7

	c.Step()
	if c.regs.FPR[1] != -5 || c.regs.PS1[1] != -7 {
		t.Fatalf("after ps_neg: FPR1=%v PS1[1]=%v, want -5/-7", c.regs.FPR[1], c.regs.PS1[1])
	}
	c.Step()
	if c.regs.FPR[3] != 5 || c.regs.PS1[3] != 7 {
		t.Fatalf("after ps_abs: FPR3=%v PS1[3]=%v, want 5/7", c.regs.FPR[3], c.regs.PS1[3])
	}
}

func TestPsMrCopiesBothSlots(t *testing.T) {
	instr := uint32(4)<<26 | 1<<21 | 0<<16 | 2<<11 | 72<<1 // ps_mr f1,f2
	c, _ := newTestCPU([]uint32{instr})
	c.regs.FPR[2], c.regs.PS1[2] = 3.25, 9.5

	c.Step()
	if c.regs.FPR[1] != 3.25 || c.regs.PS1[1] != 9.5 {
		t.Fatalf("FPR1=%v PS1[1]=%v, want 3.25/9.5", c.regs.FPR[1], c.regs.PS1[1])
	}
}

func TestPsqLoadStoreRoundTrip(t *testing.T) {
	// psq_st f1,0x100(r0) ; psq_l f2,0x100(r0)
	stq := uint32(60)<<26 | 1<<21 | 0<<16 | 0x100
	ldq := uint32(56)<<26 | 2<<21 | 0<<16 | 0x100
	c, _ := newTestCPU([]uint32{stq, ldq})
	c.regs.FPR[1], c.regs.PS1[1] = 1.5, -2.5

	c.Step()
	c.Step()
	if c.regs.FPR[2] != 1.5 {
		t.Fatalf("FPR2 = %v, want 1.5", c.regs.FPR[2])
	}
	if c.regs.PS1[2] != -2.5 {
		t.Fatalf("PS1[2] = %v, want -2.5", c.regs.PS1[2])
	}
}
