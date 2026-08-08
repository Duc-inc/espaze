package pi

import "testing"

func TestPendingRequiresBothCauseAndMask(t *testing.T) {
	p := New()
	p.SetCause(BitVI, true)
	if p.Pending() {
		t.Fatal("expected not pending: mask is 0, nothing unmasked")
	}

	p.Write32(regINTMR, 1<<BitVI)
	if !p.Pending() {
		t.Fatal("expected pending once VI's bit is both caused and unmasked")
	}
}

func TestSetCauseClearsBit(t *testing.T) {
	p := New()
	p.Write32(regINTMR, 1<<BitAI)
	p.SetCause(BitAI, true)
	if !p.Pending() {
		t.Fatal("expected pending after SetCause(true)")
	}
	p.SetCause(BitAI, false)
	if p.Pending() {
		t.Fatal("expected not pending after SetCause(false)")
	}
}

func TestReadINTSRReflectsCause(t *testing.T) {
	p := New()
	p.SetCause(BitDI, true)
	p.SetCause(BitSI, true)
	want := uint32(1<<BitDI | 1<<BitSI)
	if got := p.Read32(regINTSR); got != want {
		t.Fatalf("INTSR = %#x, want %#x", got, want)
	}
}
