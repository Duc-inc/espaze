package disc

import "testing"

func newTestImage() []byte {
	img := make([]byte, 0x1000)
	copy(img[0:6], []byte("GTEST"))
	img[magicOffset], img[magicOffset+1], img[magicOffset+2], img[magicOffset+3] = 0xC2, 0x33, 0x9F, 0x3D
	copy(img[gameNameOffset:], []byte("Test Game\x00"))
	img[dolOffsetAddr], img[dolOffsetAddr+1], img[dolOffsetAddr+2], img[dolOffsetAddr+3] = 0x00, 0x00, 0x08, 0x00
	return img
}

func TestParseHeaderReadsFields(t *testing.T) {
	h, err := ParseHeader(newTestImage())
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}
	if h.GameName != "Test Game" {
		t.Fatalf("GameName = %q, want %q", h.GameName, "Test Game")
	}
	if h.DOLOffset != 0x0800 {
		t.Fatalf("DOLOffset = %#08x, want 0x0800", h.DOLOffset)
	}
}

func TestParseHeaderRejectsBadMagic(t *testing.T) {
	img := newTestImage()
	img[magicOffset] = 0x00
	if _, err := ParseHeader(img); err == nil {
		t.Fatal("expected an error for a bad magic word")
	}
}

func TestParseHeaderRejectsTooSmallImage(t *testing.T) {
	if _, err := ParseHeader(make([]byte, 10)); err == nil {
		t.Fatal("expected an error for an image too small to hold a header")
	}
}
