package xf

import (
	"math"
	"testing"
)

func TestPosMatrixIdentityLeavesVertexUnchanged(t *testing.T) {
	v := Vec3{X: 1, Y: 2, Z: 3}
	got := IdentityPos().MulVec3(v)
	if got != v {
		t.Fatalf("got %+v, want %+v", got, v)
	}
}

func TestPosMatrixTranslation(t *testing.T) {
	m := PosMatrix{
		{1, 0, 0, 10},
		{0, 1, 0, 20},
		{0, 0, 1, 30},
	}
	got := m.MulVec3(Vec3{X: 1, Y: 1, Z: 1})
	want := Vec3{X: 11, Y: 21, Z: 31}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestPosMatrixScale(t *testing.T) {
	m := PosMatrix{
		{2, 0, 0, 0},
		{0, 3, 0, 0},
		{0, 0, 4, 0},
	}
	got := m.MulVec3(Vec3{X: 1, Y: 1, Z: 1})
	want := Vec3{X: 2, Y: 3, Z: 4}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestNormalMatrixIgnoresTranslation(t *testing.T) {
	// A NormalMatrix has no translation column at all - scaling it
	// and comparing against the equivalent PosMatrix (translation
	// zeroed) should agree.
	n := NormalMatrix{
		{2, 0, 0},
		{0, 3, 0},
		{0, 0, 4},
	}
	got := n.MulVec3(Vec3{X: 1, Y: 1, Z: 1})
	want := Vec3{X: 2, Y: 3, Z: 4}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestMat4IdentityProducesWOfOne(t *testing.T) {
	got := IdentityMat4().MulVec3(Vec3{X: 5, Y: -2, Z: 7})
	want := Vec4{X: 5, Y: -2, Z: 7, W: 1}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestMat4ProjectionProducesNonTrivialW(t *testing.T) {
	// A minimal perspective-style matrix whose bottom row derives W
	// from -Z, the way a real projection matrix would - exercising
	// the case PosMatrix/NormalMatrix can never produce.
	m := Mat4{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, -1, 0},
	}
	got := m.MulVec3(Vec3{X: 1, Y: 2, Z: 5})
	want := Vec4{X: 1, Y: 2, Z: 5, W: -5}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestVec3Dot(t *testing.T) {
	got := Vec3{X: 1, Y: 2, Z: 3}.Dot(Vec3{X: 4, Y: -5, Z: 6})
	want := float32(1*4 + 2*-5 + 3*6)
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestVec3Length(t *testing.T) {
	got := Vec3{X: 3, Y: 4, Z: 0}.Length()
	if got != 5 {
		t.Fatalf("got %v, want 5", got)
	}
}

func TestVec3NormalizeProducesUnitLength(t *testing.T) {
	got := Vec3{X: 0, Y: 0, Z: 8}.Normalize()
	want := Vec3{X: 0, Y: 0, Z: 1}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestVec3NormalizeZeroVectorReturnsUnchanged(t *testing.T) {
	v := Vec3{}
	got := v.Normalize()
	if got != v {
		t.Fatalf("got %+v, want %+v", got, v)
	}
}

func TestVec3AddSub(t *testing.T) {
	a := Vec3{X: 1, Y: 2, Z: 3}
	b := Vec3{X: 4, Y: 5, Z: 6}
	if got, want := a.Add(b), (Vec3{X: 5, Y: 7, Z: 9}); got != want {
		t.Fatalf("Add: got %+v, want %+v", got, want)
	}
	if got, want := b.Sub(a), (Vec3{X: 3, Y: 3, Z: 3}); got != want {
		t.Fatalf("Sub: got %+v, want %+v", got, want)
	}
}

func TestVec3MulScalar(t *testing.T) {
	got := Vec3{X: 1, Y: -2, Z: 3}.MulScalar(2)
	want := Vec3{X: 2, Y: -4, Z: 6}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestVec3LengthMatchesSqrtOfDotSelf(t *testing.T) {
	v := Vec3{X: 2, Y: 3, Z: 6}
	got := v.Length()
	want := float32(math.Sqrt(float64(v.Dot(v))))
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
