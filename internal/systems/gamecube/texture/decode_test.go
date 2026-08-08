package texture

import "testing"

func TestDecodeI8ProducesGrayscale(t *testing.T) {
	raw := []byte{0x80}
	got := Decode(FormatI8, raw, 1, 1)
	want := Color{R: 0x80, G: 0x80, B: 0x80, A: 255}
	if got[0] != want {
		t.Fatalf("got %+v, want %+v", got[0], want)
	}
}

func TestDecodeRGB565FullWhite(t *testing.T) {
	// 5-6-5 all-ones: 0xFFFF -> full white.
	raw := []byte{0xFF, 0xFF}
	got := Decode(FormatRGB565, raw, 1, 1)
	want := Color{R: 255, G: 255, B: 255, A: 255}
	if got[0] != want {
		t.Fatalf("got %+v, want %+v", got[0], want)
	}
}

func TestDecodeRGB565Black(t *testing.T) {
	raw := []byte{0x00, 0x00}
	got := Decode(FormatRGB565, raw, 1, 1)
	want := Color{R: 0, G: 0, B: 0, A: 255}
	if got[0] != want {
		t.Fatalf("got %+v, want %+v", got[0], want)
	}
}

func TestDecodeRGBA8PassesChannelsThrough(t *testing.T) {
	raw := []byte{10, 20, 30, 40}
	got := Decode(FormatRGBA8, raw, 1, 1)
	want := Color{R: 10, G: 20, B: 30, A: 40}
	if got[0] != want {
		t.Fatalf("got %+v, want %+v", got[0], want)
	}
}

func TestDecodeRowMajorOrder(t *testing.T) {
	// 2x1 I8 texture: first byte is (0,0), second is (1,0).
	raw := []byte{1, 2}
	got := Decode(FormatI8, raw, 2, 1)
	if got[0].R != 1 || got[1].R != 2 {
		t.Fatalf("got %+v, want row-major [1, 2]", got)
	}
}

func TestDecodeTruncatedInputLeavesRemainingTexelsZeroed(t *testing.T) {
	raw := []byte{5} // only enough for one of two texels
	got := Decode(FormatI8, raw, 2, 1)
	if got[0].R != 5 {
		t.Fatalf("got[0].R = %d, want 5", got[0].R)
	}
	if got[1] != (Color{}) {
		t.Fatalf("got[1] = %+v, want zero value", got[1])
	}
}
