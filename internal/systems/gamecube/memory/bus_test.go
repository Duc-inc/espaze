package memory

import "testing"

func TestMEM1ReadWriteRoundTrip(t *testing.T) {
	b := New()
	b.Write32(0x1000, 0xDEADBEEF)
	if v := b.Read32(0x1000); v != 0xDEADBEEF {
		t.Fatalf("Read32 = %#08x, want 0xDEADBEEF", v)
	}
}

func TestOutOfRangeReadReturnsZero(t *testing.T) {
	b := New()
	if v := b.Read8(0xFFFFFFFF); v != 0 {
		t.Fatalf("Read8(out of range) = %#02x, want 0 (unmapped)", v)
	}
}

func TestHardwareRegisterWriteDoesNotPanicOrTouchMEM1(t *testing.T) {
	b := New()
	b.Write32(hwRegBase, 0x12345678)
	if v := b.Read8(0); v != 0 {
		t.Fatalf("MEM1[0] = %#02x, want 0 (hw register write must not leak into RAM)", v)
	}
}

func TestResetClearsMEM1(t *testing.T) {
	b := New()
	b.Write8(0x10, 0xFF)
	b.Reset()
	if v := b.Read8(0x10); v != 0 {
		t.Fatalf("MEM1[0x10] after reset = %#02x, want 0", v)
	}
}
