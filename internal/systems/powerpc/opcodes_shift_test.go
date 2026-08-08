package powerpc

import "testing"

// asU32 converts through a function parameter (a non-constant value)
// so a negative literal wraps to its two's-complement bit pattern at
// runtime, rather than tripping Go's "constant overflows uint32"
// check that a direct uint32(int32(-N)) conversion hits at compile
// time.
func asU32(n int32) uint32 { return uint32(n) }

func mForm(op, rS, rA, sh, mb, me uint32, rc bool) uint32 {
	instr := op<<26 | rS<<21 | rA<<16 | sh<<11 | mb<<6 | me<<1
	if rc {
		instr |= 1
	}
	return instr
}

func TestRlwinmClearsUpperBits(t *testing.T) {
	// rlwinm r2,r1,0,16,31 - keep only the low 16 bits (SH=0, no
	// rotation, mask = IBM bits 16-31 = standard bits 0-15).
	instr := mForm(21, 1, 2, 0, 16, 31, false)
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0xFFFFFFFF

	c.Step()
	if c.regs.GPR[2] != 0x0000FFFF {
		t.Fatalf("GPR2 = %#08x, want 0x0000ffff", c.regs.GPR[2])
	}
}

func TestRlwinmRotatesBeforeMasking(t *testing.T) {
	// rlwinm r2,r1,8,0,23 - rotate left 8, keep the top 24 bits: the
	// classic "value << 8" idiom compilers emit.
	instr := mForm(21, 1, 2, 8, 0, 23, false)
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0x000000FF

	c.Step()
	if c.regs.GPR[2] != 0x0000FF00 {
		t.Fatalf("GPR2 = %#08x, want 0x0000ff00", c.regs.GPR[2])
	}
}

func TestRlwimiPreservesUnmaskedDestinationBits(t *testing.T) {
	// rlwimi r2,r1,0,16,31 - insert r1's low 16 bits into r2, leaving
	// r2's upper 16 bits exactly as they were.
	instr := mForm(20, 1, 2, 0, 16, 31, false)
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0x00000000
	c.regs.GPR[2] = 0xFFFFFFFF

	c.Step()
	if c.regs.GPR[2] != 0xFFFF0000 {
		t.Fatalf("GPR2 = %#08x, want 0xffff0000", c.regs.GPR[2])
	}
}

func TestRlwnmUsesShiftAmountFromRegister(t *testing.T) {
	// rlwnm r3,r1,r2,0,31 - rotate left by the amount in r2, full mask.
	instr := uint32(23)<<26 | 1<<21 | 3<<16 | 2<<11 | 0<<6 | 31<<1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0x00000001
	c.regs.GPR[2] = 4

	c.Step()
	if c.regs.GPR[3] != 0x00000010 {
		t.Fatalf("GPR3 = %#08x, want 0x00000010", c.regs.GPR[3])
	}
}

func TestSlwShiftsLeftAndSetsCR0WhenRequested(t *testing.T) {
	// slw. r3,r1,r2
	instr := uint32(31)<<26 | 1<<21 | 3<<16 | 2<<11 | 24<<1 | 1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 1
	c.regs.GPR[2] = 4

	c.Step()
	if c.regs.GPR[3] != 16 {
		t.Fatalf("GPR3 = %d, want 16", c.regs.GPR[3])
	}
	if c.regs.CR>>28 != 0x4 { // positive result -> "greater than"
		t.Fatalf("CR field0 = %#x, want 0x4", c.regs.CR>>28)
	}
}

func TestSrwShiftsRightLogically(t *testing.T) {
	instr := uint32(31)<<26 | 1<<21 | 3<<16 | 2<<11 | 536<<1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = 0x80000000
	c.regs.GPR[2] = 4

	c.Step()
	if c.regs.GPR[3] != 0x08000000 {
		t.Fatalf("GPR3 = %#08x, want 0x08000000 (zero-filled, not sign-extended)", c.regs.GPR[3])
	}
}

func TestSrawSignExtendsAndSetsCarryOnLostBits(t *testing.T) {
	// -3 >> 1 (arithmetic) = -2, and since -3's low bit is 1, a 1-bit
	// gets shifted out of a negative number: carry should be set.
	instr := uint32(31)<<26 | 1<<21 | 3<<16 | 2<<11 | 792<<1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = asU32(-3)
	c.regs.GPR[2] = 1

	c.Step()
	if int32(c.regs.GPR[3]) != -2 {
		t.Fatalf("GPR3 = %d, want -2", int32(c.regs.GPR[3]))
	}
	if !c.regs.getXER(XERCarry) {
		t.Fatal("expected XER carry to be set")
	}
}

func TestSrawNoCarryWhenNoBitsLost(t *testing.T) {
	// -4 >> 1 (arithmetic) = -2, with no 1-bits shifted out.
	instr := uint32(31)<<26 | 1<<21 | 3<<16 | 2<<11 | 792<<1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = asU32(-4)
	c.regs.GPR[2] = 1

	c.Step()
	if int32(c.regs.GPR[3]) != -2 {
		t.Fatalf("GPR3 = %d, want -2", int32(c.regs.GPR[3]))
	}
	if c.regs.getXER(XERCarry) {
		t.Fatal("expected XER carry to be clear")
	}
}

func TestSrawShiftOf32OrMoreSaturates(t *testing.T) {
	instr := uint32(31)<<26 | 1<<21 | 3<<16 | 2<<11 | 792<<1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = asU32(-5)
	c.regs.GPR[2] = 32

	c.Step()
	if int32(c.regs.GPR[3]) != -1 {
		t.Fatalf("GPR3 = %d, want -1", int32(c.regs.GPR[3]))
	}
	if !c.regs.getXER(XERCarry) {
		t.Fatal("expected XER carry to be set (negative source, shift >= 32)")
	}
}

func TestSrawiUsesImmediateShiftAmount(t *testing.T) {
	// srawi r3,r1,1
	instr := uint32(31)<<26 | 1<<21 | 3<<16 | 1<<11 | 824<<1
	c, _ := newTestCPU([]uint32{instr})
	c.regs.GPR[1] = asU32(-3)

	c.Step()
	if int32(c.regs.GPR[3]) != -2 {
		t.Fatalf("GPR3 = %d, want -2", int32(c.regs.GPR[3]))
	}
	if !c.regs.getXER(XERCarry) {
		t.Fatal("expected XER carry to be set")
	}
}
