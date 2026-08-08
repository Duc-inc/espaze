package exi

import "testing"

func TestCSRTSTARTSelfClears(t *testing.T) {
	e := New()
	e.Write32(regCSR, 0xFF) // includes TSTART (bit 0) plus other bits
	got := e.Read32(regCSR)
	if got&bitTSTART != 0 {
		t.Fatalf("CSR = %#x, want TSTART cleared", got)
	}
	if got != 0xFE {
		t.Fatalf("CSR = %#x, want 0xfe (every other bit preserved)", got)
	}
}

func TestDataRegisterRoundTrips(t *testing.T) {
	e := New()
	e.Write32(regData, 0xDEADBEEF)
	if got := e.Read32(regData); got != 0xDEADBEEF {
		t.Fatalf("DATA = %#08x, want 0xdeadbeef", got)
	}
}

func TestChannelsAreIndependent(t *testing.T) {
	e := New()
	e.Write32(0*channelStride+regData, 0x11111111)
	e.Write32(1*channelStride+regData, 0x22222222)
	e.Write32(2*channelStride+regData, 0x33333333)

	if got := e.Read32(0*channelStride + regData); got != 0x11111111 {
		t.Fatalf("channel 0 DATA = %#08x, want 0x11111111", got)
	}
	if got := e.Read32(1*channelStride + regData); got != 0x22222222 {
		t.Fatalf("channel 1 DATA = %#08x, want 0x22222222", got)
	}
	if got := e.Read32(2*channelStride + regData); got != 0x33333333 {
		t.Fatalf("channel 2 DATA = %#08x, want 0x33333333", got)
	}
}

func TestOutOfRangeChannelIsIgnored(t *testing.T) {
	e := New()
	e.Write32(3*channelStride+regData, 0xAAAAAAAA) // channel 3 doesn't exist
	if got := e.Read32(3*channelStride + regData); got != 0 {
		t.Fatalf("out-of-range channel read = %#08x, want 0", got)
	}
}
