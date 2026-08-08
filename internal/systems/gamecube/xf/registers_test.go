package xf

import "testing"

func TestNewRegistersDefaultViewportIsUnscaled(t *testing.T) {
	r := NewRegisters()
	want := Viewport{ScaleX: 1, ScaleY: 1, ScaleZ: 1}
	if r.Viewport != want {
		t.Fatalf("got %+v, want %+v", r.Viewport, want)
	}
}

func TestWriteOutOfRangeIsIgnored(t *testing.T) {
	var r Registers
	r.Write(RegistersStart-1, 0xABCD) // below the register block
	r.Write(RegistersEnd, 0xABCD)     // at/above the end
	if r.Raw != ([registerCount]uint32{}) {
		t.Fatalf("expected Raw to stay zeroed, got %+v", r.Raw)
	}
}

func TestWriteMatrixSelection0DecodesAllFields(t *testing.T) {
	var r Registers
	// geometry=1, tex0=2, tex1=3, tex2=4, tex3=5
	word := uint32(1) | 2<<6 | 3<<12 | 4<<18 | 5<<24
	r.Write(RegMatrixSelection0, word)

	if got := r.MatrixSelection0.GeometryIndex(); got != 1 {
		t.Fatalf("GeometryIndex = %d, want 1", got)
	}
	if got := r.MatrixSelection0.Texture0Index(); got != 2 {
		t.Fatalf("Texture0Index = %d, want 2", got)
	}
	if got := r.MatrixSelection0.Texture1Index(); got != 3 {
		t.Fatalf("Texture1Index = %d, want 3", got)
	}
	if got := r.MatrixSelection0.Texture2Index(); got != 4 {
		t.Fatalf("Texture2Index = %d, want 4", got)
	}
	if got := r.MatrixSelection0.Texture3Index(); got != 5 {
		t.Fatalf("Texture3Index = %d, want 5", got)
	}
}

func TestWriteMatrixSelection1DecodesAllFields(t *testing.T) {
	var r Registers
	word := uint32(6) | 7<<6 | 8<<12 | 9<<18
	r.Write(RegMatrixSelection1, word)

	if got := r.MatrixSelection1.Texture4Index(); got != 6 {
		t.Fatalf("Texture4Index = %d, want 6", got)
	}
	if got := r.MatrixSelection1.Texture7Index(); got != 9 {
		t.Fatalf("Texture7Index = %d, want 9", got)
	}
}

func TestWritePreservesRawWordForUndecodedRegister(t *testing.T) {
	var r Registers
	r.Write(RegVertexSpec, 0x1234)
	if r.Raw[RegVertexSpec-RegistersStart] != 0x1234 {
		t.Fatalf("Raw[RegVertexSpec] = %#x, want 0x1234", r.Raw[RegVertexSpec-RegistersStart])
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
