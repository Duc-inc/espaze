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

type fakePeripheral struct{ reg uint32 }

func (p *fakePeripheral) Read32(offset uint32) uint32       { return p.reg }
func (p *fakePeripheral) Write32(offset uint32, val uint32) { p.reg = val }

func TestAttachedPeripheralHandlesWordAccess(t *testing.T) {
	b := New()
	dev := &fakePeripheral{}
	b.Attach(hwRegBase, 0x100, dev)

	b.Write32(hwRegBase+4, 0xCAFEBABE)
	if dev.reg != 0xCAFEBABE {
		t.Fatalf("peripheral reg = %#08x, want 0xcafebabe", dev.reg)
	}
	if got := b.Read32(hwRegBase + 4); got != 0xCAFEBABE {
		t.Fatalf("Read32 = %#08x, want 0xcafebabe", got)
	}
}

func TestAttachedPeripheralHandlesByteAccess(t *testing.T) {
	b := New()
	dev := &fakePeripheral{reg: 0x11223344}
	b.Attach(hwRegBase, 0x100, dev)

	if got := b.Read8(hwRegBase); got != 0x11 {
		t.Fatalf("Read8[0] = %#02x, want 0x11", got)
	}
	if got := b.Read8(hwRegBase + 3); got != 0x44 {
		t.Fatalf("Read8[3] = %#02x, want 0x44", got)
	}
	b.Write8(hwRegBase+3, 0xFF)
	if dev.reg != 0x112233FF {
		t.Fatalf("reg after byte write = %#08x, want 0x112233ff", dev.reg)
	}
}
