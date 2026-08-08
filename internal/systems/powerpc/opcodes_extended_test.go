package powerpc

import "testing"

func xForm(op, rS, rA, rB, xo uint32, rc bool) uint32 {
	instr := op<<26 | rS<<21 | rA<<16 | rB<<11 | xo<<1
	if rc {
		instr |= 1
	}
	return instr
}

func TestSubficComputesSimmMinusRAWithCarry(t *testing.T) {
	// subfic r3,r1,10 with r1=5 -> 10-5=5, no borrow -> carry set.
	instr := uint32(8)<<26 | 3<<21 | 1<<16 | 10
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 5

	c.Step()
	if c.regs.GPR[3] != 5 {
		t.Fatalf("GPR3 = %d, want 5", c.regs.GPR[3])
	}
	if !c.regs.getXER(XERCarry) {
		t.Fatal("expected carry set (no borrow)")
	}
}

func TestSubficBorrowClearsCarry(t *testing.T) {
	// subfic r3,r1,5 with r1=10 -> 5-10=-5, a borrow occurred -> carry clear.
	instr := uint32(8)<<26 | 3<<21 | 1<<16 | 5
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 10

	c.Step()
	if int32(c.regs.GPR[3]) != -5 {
		t.Fatalf("GPR3 = %d, want -5", int32(c.regs.GPR[3]))
	}
	if c.regs.getXER(XERCarry) {
		t.Fatal("expected carry clear (borrow occurred)")
	}
}

func TestAddicSetsCarryOnOverflow(t *testing.T) {
	// addic r3,r1,1 with r1=0xFFFFFFFF -> wraps to 0, carry set.
	instr := uint32(12)<<26 | 3<<21 | 1<<16 | 1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0xFFFFFFFF

	c.Step()
	if c.regs.GPR[3] != 0 {
		t.Fatalf("GPR3 = %d, want 0", c.regs.GPR[3])
	}
	if !c.regs.getXER(XERCarry) {
		t.Fatal("expected carry set on wraparound")
	}
}

func TestNegNegatesRegister(t *testing.T) {
	instr := xForm(31, 3, 1, 0, 104, false) // neg r3,r1 (no rB)
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 5

	c.Step()
	if int32(c.regs.GPR[3]) != -5 {
		t.Fatalf("GPR3 = %d, want -5", int32(c.regs.GPR[3]))
	}
}

func TestMulhwReturnsHighWordOfSignedProduct(t *testing.T) {
	instr := xForm(31, 3, 1, 2, 75, false)
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0x00010000
	c.regs.GPR[2] = 0x00010000 // product = 0x1_0000_0000

	c.Step()
	if c.regs.GPR[3] != 1 {
		t.Fatalf("GPR3 = %d, want 1", c.regs.GPR[3])
	}
}

func TestMulhwuReturnsHighWordOfUnsignedProduct(t *testing.T) {
	instr := xForm(31, 3, 1, 2, 11, false)
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0xFFFFFFFF
	c.regs.GPR[2] = 2 // product = 0x1_FFFF_FFFE

	c.Step()
	if c.regs.GPR[3] != 1 {
		t.Fatalf("GPR3 = %d, want 1", c.regs.GPR[3])
	}
}

func TestDivwuDividesUnsigned(t *testing.T) {
	instr := xForm(31, 3, 1, 2, 459, false)
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0xFFFFFFFE // huge as unsigned, negative as signed
	c.regs.GPR[2] = 2

	c.Step()
	if c.regs.GPR[3] != 0x7FFFFFFF {
		t.Fatalf("GPR3 = %#08x, want 0x7fffffff", c.regs.GPR[3])
	}
}

func TestEqvIsBitwiseEquivalence(t *testing.T) {
	instr := xForm(31, 1, 3, 2, 284, false) // eqv r3,r1,r2
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0xFF
	c.regs.GPR[2] = 0xFF

	c.Step()
	if c.regs.GPR[3] != 0xFFFFFFFF {
		t.Fatalf("GPR3 = %#08x, want 0xffffffff", c.regs.GPR[3])
	}
}

func TestAndcClearsBitsSetInRB(t *testing.T) {
	instr := xForm(31, 1, 3, 2, 60, false) // andc r3,r1,r2
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0xFF
	c.regs.GPR[2] = 0x0F

	c.Step()
	if c.regs.GPR[3] != 0xF0 {
		t.Fatalf("GPR3 = %#08x, want 0xf0", c.regs.GPR[3])
	}
}

func TestOrcOrsWithComplementOfRB(t *testing.T) {
	instr := xForm(31, 1, 3, 2, 412, false) // orc r3,r1,r2
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0x00000000
	c.regs.GPR[2] = 0x00000000 // ^r2 = 0xffffffff

	c.Step()
	if c.regs.GPR[3] != 0xFFFFFFFF {
		t.Fatalf("GPR3 = %#08x, want 0xffffffff", c.regs.GPR[3])
	}
}

func TestExtsbSignExtendsByte(t *testing.T) {
	instr := xForm(31, 1, 3, 0, 954, false) // extsb r3,r1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0xFF

	c.Step()
	if c.regs.GPR[3] != 0xFFFFFFFF {
		t.Fatalf("GPR3 = %#08x, want 0xffffffff", c.regs.GPR[3])
	}
}

func TestExtshSignExtendsHalfword(t *testing.T) {
	instr := xForm(31, 1, 3, 0, 922, false) // extsh r3,r1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0xFFFF

	c.Step()
	if c.regs.GPR[3] != 0xFFFFFFFF {
		t.Fatalf("GPR3 = %#08x, want 0xffffffff", c.regs.GPR[3])
	}
}

func TestLhaSignExtendsLoadedHalfword(t *testing.T) {
	// stw storing 0xFFFF as the low halfword at 0x100 ; lha r2,0x102(r0)
	stw := uint32(36)<<26 | 1<<21 | 0<<16 | 0x100
	lha := uint32(42)<<26 | 2<<21 | 0<<16 | 0x102
	c, _ := newTestCPU([]uint32{stw, lha})
	c.regs.GPR[1] = 0x0000FFFF

	c.Step()
	c.Step()
	if c.regs.GPR[2] != 0xFFFFFFFF {
		t.Fatalf("GPR2 = %#08x, want 0xffffffff", c.regs.GPR[2])
	}
}

func TestIndexedLoadStoreRoundTrip(t *testing.T) {
	// r1=0x100 (base), r2=4 (index), r3=0xABCDEF01 (value)
	// stwx r3,r1,r2 ; lwzx r4,r1,r2
	stwx := xForm(31, 3, 1, 2, 151, false)
	lwzx := xForm(31, 4, 1, 2, 23, false)
	c, _ := newTestCPU([]uint32{stwx, lwzx})
	c.regs.GPR[1] = 0x100
	c.regs.GPR[2] = 4
	c.regs.GPR[3] = 0xABCDEF01

	c.Step()
	c.Step()
	if c.regs.GPR[4] != 0xABCDEF01 {
		t.Fatalf("GPR4 = %#08x, want 0xabcdef01", c.regs.GPR[4])
	}
}

func TestLmwLoadsConsecutiveRegisters(t *testing.T) {
	// lmw r29,0x200(r0) - loads r29, r30, r31 from three consecutive
	// words starting at 0x200.
	base := uint32(0x200)
	lmw := uint32(46)<<26 | 29<<21 | 0<<16 | base
	c, bus := newTestCPU([]uint32{lmw})
	bus.Write32(base, 0x11)
	bus.Write32(base+4, 0x22)
	bus.Write32(base+8, 0x33)

	c.Step()

	if c.regs.GPR[29] != 0x11 || c.regs.GPR[30] != 0x22 || c.regs.GPR[31] != 0x33 {
		t.Fatalf("GPR29-31 = %d,%d,%d, want 0x11,0x22,0x33", c.regs.GPR[29], c.regs.GPR[30], c.regs.GPR[31])
	}
}

func TestStmwStoresConsecutiveRegisters(t *testing.T) {
	stmw := uint32(47)<<26 | 29<<21 | 0<<16 | 0x200
	c, bus := newTestCPU([]uint32{stmw})
	c.regs.GPR[29] = 0xAA
	c.regs.GPR[30] = 0xBB
	c.regs.GPR[31] = 0xCC

	c.Step()

	if bus.Read32(0x200) != 0xAA || bus.Read32(0x204) != 0xBB || bus.Read32(0x208) != 0xCC {
		t.Fatalf("stored words = %#x,%#x,%#x, want 0xaa,0xbb,0xcc", bus.Read32(0x200), bus.Read32(0x204), bus.Read32(0x208))
	}
}

func TestAddcAddeCarryChain(t *testing.T) {
	// addc r3,r1,r2: 0xFFFFFFFF + 2 overflows 32 bits -> carry set.
	addc := xForm(31, 3, 1, 2, 10, false)
	// adde r4,r5,r6 + carry-in from addc above.
	adde := xForm(31, 4, 5, 6, 138, false)
	c, _ := newTestCPU([]uint32{addc, adde})
	c.regs.GPR[1] = 0xFFFFFFFF
	c.regs.GPR[2] = 2
	c.regs.GPR[5] = 10
	c.regs.GPR[6] = 20

	c.Step()
	if c.regs.GPR[3] != 1 {
		t.Fatalf("GPR3 (addc) = %d, want 1", c.regs.GPR[3])
	}
	if !c.regs.getXER(XERCarry) {
		t.Fatal("expected carry set after addc overflow")
	}
	c.Step()
	if c.regs.GPR[4] != 31 { // 10+20+1(carry-in)
		t.Fatalf("GPR4 (adde) = %d, want 31", c.regs.GPR[4])
	}
}

func TestSubfcSubfeBorrowChain(t *testing.T) {
	// subfc r3,r1,r2: rB-rA = 5-10 = -5 -> borrow -> carry clear.
	subfc := xForm(31, 3, 1, 2, 8, false)
	subfe := xForm(31, 4, 5, 6, 136, false)
	c, _ := newTestCPU([]uint32{subfc, subfe})
	c.regs.GPR[1] = 10
	c.regs.GPR[2] = 5
	c.regs.GPR[5] = 3
	c.regs.GPR[6] = 10

	c.Step()
	if int32(c.regs.GPR[3]) != -5 {
		t.Fatalf("GPR3 (subfc) = %d, want -5", int32(c.regs.GPR[3]))
	}
	if c.regs.getXER(XERCarry) {
		t.Fatal("expected carry clear after subfc borrow")
	}
	c.Step()
	if int32(c.regs.GPR[4]) != 6 { // 10-3-1(no carry-in, i.e. borrow propagated)
		t.Fatalf("GPR4 (subfe) = %d, want 6", int32(c.regs.GPR[4]))
	}
}

func TestCntlzwCountsLeadingZeros(t *testing.T) {
	cntlzw := xForm(31, 1, 3, 0, 26, false)
	c, _ := newTestCPU([]uint32{cntlzw})
	c.regs.GPR[1] = 0x0000_00F0 // 24 leading zeros

	c.Step()
	if c.regs.GPR[3] != 24 {
		t.Fatalf("GPR3 (cntlzw) = %d, want 24", c.regs.GPR[3])
	}
}

func TestMtcrfUpdatesOnlySelectedFields(t *testing.T) {
	// mtcrf 0x80,r1: FXM=0x80 selects only CR field 0 (the top 4 bits).
	instr := uint32(31)<<26 | 1<<21 | 0x80<<12 | 144<<1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0xFFFFFFFF
	c.regs.CR = 0x0000000F // field 7 pre-set, should survive untouched

	c.Step()
	if c.regs.CR != 0xF000000F {
		t.Fatalf("CR = %#08x, want 0xf000000f (only field 0 updated)", c.regs.CR)
	}
}

func TestMemoryBarriersAreHarmlessNoOps(t *testing.T) {
	sync := xForm(31, 0, 0, 0, 598, false)
	eieio := xForm(31, 0, 0, 0, 854, false)
	isync := uint32(19)<<26 | 0<<21 | 0<<16 | 0<<11 | 150<<1
	c, _ := newTestCPU([]uint32{sync, eieio, isync})
	c.regs.GPR[1] = 0x1234 // sentinel, must survive untouched

	c.Step()
	c.Step()
	c.Step()

	if c.regs.GPR[1] != 0x1234 {
		t.Fatalf("GPR1 = %#08x, want 0x1234 unchanged", c.regs.GPR[1])
	}
	if c.PC() != 12 {
		t.Fatalf("PC = %d, want 12 (three instructions executed)", c.PC())
	}
}
