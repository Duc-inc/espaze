package xf

import (
	"math"
	"testing"
)

func writePosMatrix(mem *Memory, addr uint16, m PosMatrix) {
	i := uint16(0)
	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			mem.Write(addr+i, math.Float32bits(m[row][col]))
			i++
		}
	}
}

func writeNormalMatrix(mem *Memory, addr uint16, m NormalMatrix) {
	i := uint16(0)
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			mem.Write(addr+i, math.Float32bits(m[row][col]))
			i++
		}
	}
}

func TestPerspectiveDivide(t *testing.T) {
	got := PerspectiveDivide(Vec4{X: 10, Y: 20, Z: 30, W: 2})
	want := Vec3{X: 5, Y: 10, Z: 15}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestPerspectiveDivideZeroWLeavesUnchanged(t *testing.T) {
	got := PerspectiveDivide(Vec4{X: 1, Y: 2, Z: 3, W: 0})
	want := Vec3{X: 1, Y: 2, Z: 3}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestViewportApply(t *testing.T) {
	vp := Viewport{ScaleX: 320, ScaleY: 240, OffsetX: 320, OffsetY: 240}
	got := vp.Apply(Vec3{X: 1, Y: -1, Z: 0.5})
	want := Vec3{X: 640, Y: 0, Z: 0.5}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestViewSpacePositionAppliesOnlyPositionMatrix(t *testing.T) {
	mem := NewMemory()
	writePosMatrix(mem, 0, PosMatrix{
		{1, 0, 0, 5},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
	})
	regs := NewRegisters()

	got := ViewSpacePosition(Vec3{X: 1, Y: 2, Z: 3}, mem, regs)
	want := Vec3{X: 6, Y: 2, Z: 3}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestTransformPositionIdentityPipelineIsUnchanged(t *testing.T) {
	mem := NewMemory()
	writePosMatrix(mem, 0, IdentityPos())

	regs := NewRegisters()
	regs.PosMatrixIndex = 0
	regs.Projection = Projection{
		Coeffs: [6]float32{1, 0, 1, 0, 1, 0},
		Type:   ProjectionOrthographic,
	}

	pos := Vec3{X: 3, Y: -4, Z: 5}
	got := TransformPosition(pos, mem, regs)
	if got != pos {
		t.Fatalf("got %+v, want %+v", got, pos)
	}
}

func TestTransformPositionAppliesTranslationAndViewport(t *testing.T) {
	mem := NewMemory()
	writePosMatrix(mem, 0, PosMatrix{
		{1, 0, 0, 10},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
	})

	regs := NewRegisters()
	regs.Projection = Projection{
		Coeffs: [6]float32{1, 0, 1, 0, 1, 0},
		Type:   ProjectionOrthographic,
	}
	regs.Viewport = Viewport{ScaleX: 1, ScaleY: 1, OffsetX: 100}

	got := TransformPosition(Vec3{X: 0, Y: 0, Z: 0}, mem, regs)
	want := Vec3{X: 110, Y: 0, Z: 0}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestTransformNormalNormalizesAfterScale(t *testing.T) {
	mem := NewMemory()
	writeNormalMatrix(mem, 0, NormalMatrix{
		{2, 0, 0},
		{0, 2, 0},
		{0, 0, 2},
	})
	regs := NewRegisters()
	regs.NormalMatrixIndex = 0

	got := TransformNormal(Vec3{X: 0, Y: 0, Z: 3}, mem, regs)
	want := Vec3{X: 0, Y: 0, Z: 1}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
