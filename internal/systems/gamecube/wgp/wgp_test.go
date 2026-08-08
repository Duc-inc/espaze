package wgp

import "testing"

func TestNoBurstUntil32Bytes(t *testing.T) {
	w := New()
	for i := 0; i < 31; i++ {
		w.StreamByte(byte(i))
	}
	if bursts := w.DrainBursts(); len(bursts) != 0 {
		t.Fatalf("got %d bursts, want 0 before 32 bytes", len(bursts))
	}
}

func TestBurstFiresAt32Bytes(t *testing.T) {
	w := New()
	for i := 0; i < 32; i++ {
		w.StreamByte(byte(i))
	}
	bursts := w.DrainBursts()
	if len(bursts) != 1 {
		t.Fatalf("got %d bursts, want 1", len(bursts))
	}
	if bursts[0][0] != 0 || bursts[0][31] != 31 {
		t.Fatalf("burst = %v, want bytes 0..31 in order", bursts[0])
	}
}

func TestPartialTailStaysBufferedAcrossDrains(t *testing.T) {
	w := New()
	for i := 0; i < 40; i++ {
		w.StreamByte(byte(i))
	}
	first := w.DrainBursts()
	if len(first) != 1 {
		t.Fatalf("got %d bursts, want 1 (8 bytes should remain buffered)", len(first))
	}
	for i := 0; i < 24; i++ {
		w.StreamByte(byte(i))
	}
	second := w.DrainBursts()
	if len(second) != 1 {
		t.Fatalf("got %d bursts, want 1 (8 leftover + 24 new = 32)", len(second))
	}
}

func TestWrite32GathersFourBytesBigEndian(t *testing.T) {
	w := New()
	for i := 0; i < 8; i++ {
		w.Write32(0, 0x01020304)
	}
	bursts := w.DrainBursts()
	if len(bursts) != 1 {
		t.Fatalf("got %d bursts, want 1", len(bursts))
	}
	if bursts[0][0] != 0x01 || bursts[0][1] != 0x02 || bursts[0][2] != 0x03 || bursts[0][3] != 0x04 {
		t.Fatalf("first 4 bytes = %v, want [1 2 3 4]", bursts[0][:4])
	}
}
