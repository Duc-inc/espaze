package gpu

import "testing"

func TestCombineReplace(t *testing.T) {
	got := Combine(TEVReplace, Color{10, 20, 30, 40}, Color{100, 100, 100, 100})
	if got != (Color{10, 20, 30, 40}) {
		t.Fatalf("Combine(Replace) = %+v, want the texel unchanged", got)
	}
}

func TestCombineModulate(t *testing.T) {
	got := Combine(TEVModulate, Color{255, 255, 255, 255}, Color{128, 0, 255, 255})
	if got.R != 128 || got.G != 0 || got.B != 255 {
		t.Fatalf("Combine(Modulate) with a white texel should pass the incoming color through, got %+v", got)
	}
}

func TestCombineAddClamps(t *testing.T) {
	got := Combine(TEVAdd, Color{200, 0, 0, 0}, Color{200, 0, 0, 0})
	if got.R != 255 {
		t.Fatalf("Combine(Add) R = %d, want clamped to 255", got.R)
	}
}

func TestCombineDecalUsesTexelAlpha(t *testing.T) {
	got := Combine(TEVDecal, Color{255, 0, 0, 255}, Color{0, 255, 0, 255})
	if got.R != 255 || got.G != 0 {
		t.Fatalf("Combine(Decal) with full texel alpha should be fully opaque texel color, got %+v", got)
	}
}
