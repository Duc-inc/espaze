package xf

import "testing"

func TestNewRegistersDefaultViewportIsUnscaled(t *testing.T) {
	r := NewRegisters()
	want := Viewport{ScaleX: 1, ScaleY: 1}
	if r.Viewport != want {
		t.Fatalf("got %+v, want %+v", r.Viewport, want)
	}
}

func TestNewRegistersDefaultHasNoTexGens(t *testing.T) {
	r := NewRegisters()
	if r.TexGenCount != 0 {
		t.Fatalf("got %d, want 0", r.TexGenCount)
	}
}

func TestProjectionMatrixOrthographic(t *testing.T) {
	p := Projection{
		Coeffs: [6]float32{2, 0, 3, 0, 4, 0},
		Type:   ProjectionOrthographic,
	}
	got := p.Matrix().MulVec3(Vec3{X: 1, Y: 1, Z: 1})
	want := Vec4{X: 2, Y: 3, Z: 4, W: 1}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestProjectionMatrixPerspectiveProducesWFromZ(t *testing.T) {
	p := Projection{
		Coeffs: [6]float32{1, 0, 1, 0, 1, 0},
		Type:   ProjectionPerspective,
	}
	got := p.Matrix().MulVec3(Vec3{X: 2, Y: 3, Z: 5})
	if got.W != -5 {
		t.Fatalf("W = %v, want -5", got.W)
	}
}
