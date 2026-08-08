package si

import "testing"

func TestReadReflectsSetChannelData(t *testing.T) {
	s := New()
	s.SetChannel(0, Channel{Connected: true, Data: [8]byte{0x3F, 0x22, 0x11, 0, 0x80, 0x81, 0x10, 0xEF}})

	high := s.Read32(regInBufHigh)
	if high&(1<<31) != 0 {
		t.Fatal("expected ERRSTAT clear for a connected channel")
	}
	if high != 0x003F2211 {
		t.Fatalf("high = %#08x, want 0x003f2211", high)
	}

	low := s.Read32(regInBufLow)
	if low != 0x808110EF {
		t.Fatalf("low = %#08x, want 0x808110ef", low)
	}
}

func TestDisconnectedChannelSetsErrStat(t *testing.T) {
	s := New()
	s.SetChannel(1, Channel{Connected: false})
	if s.Read32(channelStride+regInBufHigh)&(1<<31) == 0 {
		t.Fatal("expected ERRSTAT set for a disconnected channel")
	}
}

func TestOutOfRangeChannelIgnored(t *testing.T) {
	s := New()
	s.SetChannel(99, Channel{Connected: true}) // must not panic
	if got := s.Read32(99 * channelStride); got != 0 {
		t.Fatalf("got %#x, want 0", got)
	}
}
