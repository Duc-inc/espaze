package powerpc

import (
	"math"
	"testing"
)

func TestFloatLoadStoreDoubleRoundTrip(t *testing.T) {
	// stfd f1,0x100(r0) requires f1 preloaded via direct register poke
	// (no float-immediate-load instruction exists on real hardware
	// either - constants are always loaded from memory).
	stfd := uint32(54)<<26 | 1<<21 | 0<<16 | 0x100
	lfd := uint32(50)<<26 | 2<<21 | 0<<16 | 0x100
	c, _ := newTestCPU([]uint32{stfd, lfd})
	c.regs.FPR[1] = 3.5

	c.Step()
	c.Step()
	if c.regs.FPR[2] != 3.5 {
		t.Fatalf("FPR2 = %v, want 3.5", c.regs.FPR[2])
	}
}

func TestFloatAdd(t *testing.T) {
	// fadd f3,f1,f2 (ext63A=21)
	fadd := uint32(63)<<26 | 3<<21 | 1<<16 | 2<<11 | 21<<1
	c, _ := newTestCPU([]uint32{fadd})
	c.regs.FPR[1] = 1.5
	c.regs.FPR[2] = 2.25

	c.Step()
	if c.regs.FPR[3] != 3.75 {
		t.Fatalf("FPR3 = %v, want 3.75", c.regs.FPR[3])
	}
}

func TestFloatMulUsesFRC(t *testing.T) {
	// fmul f3,f1,f2 (ext63A=25): frA=1, frC=2 (frB field unused/zero)
	fmul := uint32(63)<<26 | 3<<21 | 1<<16 | 0<<11 | 2<<6 | 25<<1
	c, _ := newTestCPU([]uint32{fmul})
	c.regs.FPR[1] = 4.0
	c.regs.FPR[2] = 2.5

	c.Step()
	if c.regs.FPR[3] != 10.0 {
		t.Fatalf("FPR3 = %v, want 10.0", c.regs.FPR[3])
	}
}

func TestFloatNegAndAbs(t *testing.T) {
	fneg := uint32(63)<<26 | 1<<21 | 0<<16 | 2<<11 | 40<<1
	fabs := uint32(63)<<26 | 3<<21 | 0<<16 | 1<<11 | 264<<1
	c, _ := newTestCPU([]uint32{fneg, fabs})
	c.regs.FPR[2] = 5.0

	c.Step()
	if c.regs.FPR[1] != -5.0 {
		t.Fatalf("FPR1 after fneg = %v, want -5.0", c.regs.FPR[1])
	}
	c.Step()
	if c.regs.FPR[3] != 5.0 {
		t.Fatalf("FPR3 after fabs = %v, want 5.0", c.regs.FPR[3])
	}
}

func TestSinglePrecisionArithmetic(t *testing.T) {
	fadds := uint32(59)<<26 | 3<<21 | 1<<16 | 2<<11 | 21<<1
	fsubs := uint32(59)<<26 | 4<<21 | 1<<16 | 2<<11 | 20<<1
	fmuls := uint32(59)<<26 | 5<<21 | 1<<16 | 0<<11 | 2<<6 | 25<<1
	fdivs := uint32(59)<<26 | 6<<21 | 1<<16 | 2<<11 | 18<<1
	c, _ := newTestCPU([]uint32{fadds, fsubs, fmuls, fdivs})
	c.regs.FPR[1] = 5.0
	c.regs.FPR[2] = 2.0

	c.Step()
	if c.regs.FPR[3] != 7.0 {
		t.Fatalf("FPR3 (fadds) = %v, want 7.0", c.regs.FPR[3])
	}
	c.Step()
	if c.regs.FPR[4] != 3.0 {
		t.Fatalf("FPR4 (fsubs) = %v, want 3.0", c.regs.FPR[4])
	}
	c.Step()
	if c.regs.FPR[5] != 10.0 {
		t.Fatalf("FPR5 (fmuls) = %v, want 10.0", c.regs.FPR[5])
	}
	c.Step()
	if c.regs.FPR[6] != 2.5 {
		t.Fatalf("FPR6 (fdivs) = %v, want 2.5", c.regs.FPR[6])
	}
}

func TestSinglePrecisionArithmeticRoundsToFloat32(t *testing.T) {
	// fadds f3,f1,f2: 1/3 + 1/3 in double precision has more mantissa
	// bits than a float32 can hold, so the single-precision result must
	// differ from the plain double-precision sum.
	fadds := uint32(59)<<26 | 3<<21 | 1<<16 | 2<<11 | 21<<1
	c, _ := newTestCPU([]uint32{fadds})
	c.regs.FPR[1] = 1.0 / 3.0
	c.regs.FPR[2] = 1.0 / 3.0

	c.Step()

	want := float64(float32(1.0/3.0) + float32(1.0/3.0))
	if c.regs.FPR[3] != want {
		t.Fatalf("FPR3 = %v, want %v (rounded through float32)", c.regs.FPR[3], want)
	}
	doubleSum := 1.0/3.0 + 1.0/3.0
	if c.regs.FPR[3] == doubleSum {
		t.Fatal("expected single-precision result to differ from the plain double-precision sum")
	}
}

func TestFusedMultiplyAddFamily(t *testing.T) {
	// fmadd/fmsub/fnmadd/fnmsub f_,f1,f2(frC),f3(frB) - A-form field
	// order is frD(21) frA(16) frB(11) frC(6) xo(1).
	fmadd := uint32(63)<<26 | 4<<21 | 1<<16 | 3<<11 | 2<<6 | 29<<1
	fmsub := uint32(63)<<26 | 5<<21 | 1<<16 | 3<<11 | 2<<6 | 28<<1
	fnmadd := uint32(63)<<26 | 6<<21 | 1<<16 | 3<<11 | 2<<6 | 31<<1
	fnmsub := uint32(63)<<26 | 7<<21 | 1<<16 | 3<<11 | 2<<6 | 30<<1
	c, _ := newTestCPU([]uint32{fmadd, fmsub, fnmadd, fnmsub})
	c.regs.FPR[1] = 2.0 // frA
	c.regs.FPR[2] = 3.0 // frC
	c.regs.FPR[3] = 1.0 // frB

	c.Step()
	if c.regs.FPR[4] != 7.0 { // 2*3+1
		t.Fatalf("FPR4 (fmadd) = %v, want 7.0", c.regs.FPR[4])
	}
	c.Step()
	if c.regs.FPR[5] != 5.0 { // 2*3-1
		t.Fatalf("FPR5 (fmsub) = %v, want 5.0", c.regs.FPR[5])
	}
	c.Step()
	if c.regs.FPR[6] != -7.0 { // -(2*3+1)
		t.Fatalf("FPR6 (fnmadd) = %v, want -7.0", c.regs.FPR[6])
	}
	c.Step()
	if c.regs.FPR[7] != -5.0 { // -(2*3-1)
		t.Fatalf("FPR7 (fnmsub) = %v, want -5.0", c.regs.FPR[7])
	}
}

func TestFusedMultiplyAddSinglePrecisionFamily(t *testing.T) {
	fmadds := uint32(59)<<26 | 4<<21 | 1<<16 | 3<<11 | 2<<6 | 29<<1
	c, _ := newTestCPU([]uint32{fmadds})
	c.regs.FPR[1] = 2.0
	c.regs.FPR[2] = 3.0
	c.regs.FPR[3] = 1.0

	c.Step()
	if c.regs.FPR[4] != 7.0 {
		t.Fatalf("FPR4 (fmadds) = %v, want 7.0", c.regs.FPR[4])
	}
}

func TestFrspRoundsToSinglePrecision(t *testing.T) {
	frsp := uint32(63)<<26 | 2<<21 | 0<<16 | 1<<11 | 12<<1
	c, _ := newTestCPU([]uint32{frsp})
	c.regs.FPR[1] = 1.0 / 3.0

	c.Step()

	want := float64(float32(1.0 / 3.0))
	if c.regs.FPR[2] != want {
		t.Fatalf("FPR2 = %v, want %v (rounded through float32)", c.regs.FPR[2], want)
	}
}

func TestFselPicksBasedOnSign(t *testing.T) {
	// fsel f4,f1,f3,f2: frA=1(test), frC=3(if>=0), frB=2(if<0).
	fsel := uint32(63)<<26 | 4<<21 | 1<<16 | 2<<11 | 3<<6 | 23<<1
	c, _ := newTestCPU([]uint32{fsel, fsel})
	c.regs.FPR[2] = 100 // frB
	c.regs.FPR[3] = 200 // frC

	c.regs.FPR[1] = 1.0 // frA >= 0
	c.Step()
	if c.regs.FPR[4] != 200 {
		t.Fatalf("FPR4 (frA>=0) = %v, want 200 (frC)", c.regs.FPR[4])
	}

	c.regs.FPR[1] = -1.0 // frA < 0
	c.regs.PC = 0        // re-run the same instruction
	c.Step()
	if c.regs.FPR[4] != 100 {
		t.Fatalf("FPR4 (frA<0) = %v, want 100 (frB)", c.regs.FPR[4])
	}
}

func TestFctiwzTruncatesTowardZero(t *testing.T) {
	fctiwz := uint32(63)<<26 | 2<<21 | 0<<16 | 1<<11 | 15<<1
	c, _ := newTestCPU([]uint32{fctiwz})
	c.regs.FPR[1] = 3.9

	c.Step()

	got := int32(math.Float64bits(c.regs.FPR[2]))
	if got != 3 {
		t.Fatalf("truncated int = %d, want 3", got)
	}
}

func TestFctiwzClampsOutOfRangeValues(t *testing.T) {
	fctiwz := uint32(63)<<26 | 2<<21 | 0<<16 | 1<<11 | 15<<1
	c, _ := newTestCPU([]uint32{fctiwz})
	c.regs.FPR[1] = 1e20

	c.Step()

	got := int32(math.Float64bits(c.regs.FPR[2]))
	if got != 2147483647 {
		t.Fatalf("clamped int = %d, want 2147483647", got)
	}
}

func TestFloatCompareSetsConditionRegister(t *testing.T) {
	// fcmpu cr0,f1,f2 (ext63=0)
	fcmpu := uint32(63)<<26 | 0<<21 | 1<<16 | 2<<11
	c, _ := newTestCPU([]uint32{fcmpu})
	c.regs.FPR[1] = 1.0
	c.regs.FPR[2] = 2.0

	c.Step()
	if c.regs.CR>>28 != 0x8 { // "less than"
		t.Fatalf("CR field0 = %#x, want 0x8 (less than)", c.regs.CR>>28)
	}
}
