package vi

import "testing"

func TestWriteDCREnablesAndSetsFormat(t *testing.T) {
	v := New()
	v.Write32(regDCR, 1|2<<8) // ENB=1, FMT=PAL(2)
	if !v.Enabled() {
		t.Fatal("expected VI enabled")
	}
	if v.format != 2 {
		t.Fatalf("format = %d, want 2", v.format)
	}
}

func TestWriteTFBLSetsFramebufferAddr(t *testing.T) {
	v := New()
	v.Write32(regTFBL, 0x00435A00)
	if got := v.FramebufferAddr(); got != 0x00435A00 {
		t.Fatalf("FramebufferAddr = %#x, want 0x00435a00", got)
	}
}

func TestStepFiresEnabledDisplayInterruptAtItsLine(t *testing.T) {
	v := New()
	v.Write32(regDI0, 1<<28|10) // enabled, line 10

	var fired bool
	for i := 0; i < 10; i++ {
		fired = v.Step()
	}
	if !fired {
		t.Fatal("expected the interrupt to fire at line 10")
	}
	if v.Read32(regDI0)&(1<<31) == 0 {
		t.Fatal("expected INT status bit set")
	}
}

func TestStepDoesNotFireDisabledInterrupt(t *testing.T) {
	v := New()
	v.Write32(regDI0, 10) // ENB bit clear

	for i := 0; i < 10; i++ {
		if v.Step() {
			t.Fatal("expected no interrupt while disabled")
		}
	}
}

func TestWritingZeroToINTClearsIt(t *testing.T) {
	v := New()
	v.Write32(regDI0, 1<<28|5)
	for i := 0; i < 5; i++ {
		v.Step()
	}
	if v.Read32(regDI0)&(1<<31) == 0 {
		t.Fatal("expected INT set before clearing")
	}
	v.Write32(regDI0, 1<<28|5) // INT bit 31 left clear -> clears status
	if v.Read32(regDI0)&(1<<31) != 0 {
		t.Fatal("expected INT cleared after write")
	}
}

func TestStepWrapsRasterPositionAtFrameEnd(t *testing.T) {
	v := New()
	for i := 0; i < linesPerFrame; i++ {
		v.Step()
	}
	if v.vpos != 1 {
		t.Fatalf("vpos = %d, want wrapped to 1", v.vpos)
	}
}
