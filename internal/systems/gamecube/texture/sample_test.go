package texture

import "testing"

func TestSampleReturnsCorrectTexel(t *testing.T) {
	tex := New(FormatI8, []byte{1, 2, 3, 4}, 2, 2)
	if got := tex.Sample(1, 0); got.R != 2 {
		t.Fatalf("Sample(1,0).R = %d, want 2", got.R)
	}
	if got := tex.Sample(0, 1); got.R != 3 {
		t.Fatalf("Sample(0,1).R = %d, want 3", got.R)
	}
}

func TestSampleWrapsCoordinates(t *testing.T) {
	tex := New(FormatI8, []byte{1, 2, 3, 4}, 2, 2)
	if got := tex.Sample(2, 0); got.R != 1 {
		t.Fatalf("Sample(2,0).R = %d, want 1 (wrapped)", got.R)
	}
	if got := tex.Sample(-1, 0); got.R != 2 {
		t.Fatalf("Sample(-1,0).R = %d, want 2 (wrapped)", got.R)
	}
}

func TestSampleZeroSizeTextureReturnsZeroColor(t *testing.T) {
	tex := &Texture{}
	if got := tex.Sample(0, 0); got != (Color{}) {
		t.Fatalf("got %+v, want zero value", got)
	}
}
