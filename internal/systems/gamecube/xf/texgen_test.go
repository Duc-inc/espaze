package xf

import "testing"

func TestTexGenDecodesAllConfirmedFields(t *testing.T) {
	// EmbossLight=3, EmbossSource=5, SourceRow=SourceTex2(7), Type=TexGenEmbossMap(1)
	raw := uint32(3)<<15 | 5<<12 | 7<<7 | 1<<4
	tg := TexGen{Raw: raw}

	if got := tg.EmbossLight(); got != 3 {
		t.Fatalf("EmbossLight = %d, want 3", got)
	}
	if got := tg.EmbossSource(); got != 5 {
		t.Fatalf("EmbossSource = %d, want 5", got)
	}
	if got := tg.SourceRow(); got != SourceTex2 {
		t.Fatalf("SourceRow = %d, want SourceTex2", got)
	}
	if got := tg.Type(); got != TexGenEmbossMap {
		t.Fatalf("Type = %d, want TexGenEmbossMap", got)
	}
}

func TestPostTexGenDecodesConfirmedFields(t *testing.T) {
	raw := uint32(1)<<8 | 42 // NormalizeBeforeTransform set, MatrixRow=42
	ptg := PostTexGen{Raw: raw}

	if !ptg.NormalizeBeforeTransform() {
		t.Fatal("expected NormalizeBeforeTransform to be true")
	}
	if got := ptg.MatrixRow(); got != 42 {
		t.Fatalf("MatrixRow = %d, want 42", got)
	}
}

func TestRegistersWriteDecodesTexGenAndPostTexGen(t *testing.T) {
	var r Registers
	r.Write(RegTexCoordCtrlStart+2, 1<<4) // TEX2: Type = TexGenEmbossMap
	r.Write(RegPostTexCtrlStart+3, 10)    // DUALTEX3: MatrixRow = 10

	if got := r.TexGen[2].Type(); got != TexGenEmbossMap {
		t.Fatalf("TexGen[2].Type = %d, want TexGenEmbossMap", got)
	}
	if got := r.PostTexGen[3].MatrixRow(); got != 10 {
		t.Fatalf("PostTexGen[3].MatrixRow = %d, want 10", got)
	}
}
