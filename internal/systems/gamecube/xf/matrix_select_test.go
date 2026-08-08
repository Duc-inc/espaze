package xf

import "testing"

func TestPositionAndNormalMatrixAddrShareGeometryIndex(t *testing.T) {
	var regs Registers
	regs.Write(RegMatrixSelection0, 5) // GeometryIndex = 5

	if got, want := regs.PositionMatrixAddr(), uint16(PosMatricesStart+5*4); got != want {
		t.Fatalf("PositionMatrixAddr = %#x, want %#x", got, want)
	}
	if got, want := regs.NormalMatrixAddr(), uint16(NormalMatricesStart+5*3); got != want {
		t.Fatalf("NormalMatrixAddr = %#x, want %#x", got, want)
	}
}

func TestTextureMatrixAddrUsesPerSlotIndex(t *testing.T) {
	var regs Registers
	regs.Write(RegMatrixSelection0, 3<<6) // Texture0Index = 3
	regs.Write(RegMatrixSelection1, 7)    // Texture4Index = 7

	if got, want := regs.TextureMatrixAddr(0), uint16(PosMatricesStart+3*4); got != want {
		t.Fatalf("TextureMatrixAddr(0) = %#x, want %#x", got, want)
	}
	if got, want := regs.TextureMatrixAddr(4), uint16(PosMatricesStart+7*4); got != want {
		t.Fatalf("TextureMatrixAddr(4) = %#x, want %#x", got, want)
	}
}

func TestMatrixAddrsDefaultToRowZero(t *testing.T) {
	var regs Registers
	if regs.PositionMatrixAddr() != PosMatricesStart {
		t.Fatalf("PositionMatrixAddr = %#x, want %#x", regs.PositionMatrixAddr(), uint16(PosMatricesStart))
	}
	if regs.NormalMatrixAddr() != NormalMatricesStart {
		t.Fatalf("NormalMatrixAddr = %#x, want %#x", regs.NormalMatrixAddr(), uint16(NormalMatricesStart))
	}
}
