package xf

import (
	"math"
	"testing"
)

func TestMemoryWriteReadFloat32RoundTrip(t *testing.T) {
	m := NewMemory()
	m.Write(10, math.Float32bits(3.5))
	if got := m.ReadFloat32(10); got != 3.5 {
		t.Fatalf("got %v, want 3.5", got)
	}
}

func TestMemoryOutOfRangeWriteIgnored(t *testing.T) {
	m := NewMemory()
	m.Write(memSize+100, math.Float32bits(1.0)) // must not panic
}

func TestMemoryOutOfRangeReadReturnsZero(t *testing.T) {
	m := NewMemory()
	if got := m.ReadFloat32(memSize + 100); got != 0 {
		t.Fatalf("got %v, want 0", got)
	}
}

func TestMemoryReadPosMatrix(t *testing.T) {
	m := NewMemory()
	want := PosMatrix{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
	}
	var addr uint16 = 100
	i := uint16(0)
	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			m.Write(addr+i, math.Float32bits(want[row][col]))
			i++
		}
	}
	got := m.ReadPosMatrix(addr)
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestMemoryReadNormalMatrix(t *testing.T) {
	m := NewMemory()
	want := NormalMatrix{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}
	var addr uint16 = 200
	i := uint16(0)
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			m.Write(addr+i, math.Float32bits(want[row][col]))
			i++
		}
	}
	got := m.ReadNormalMatrix(addr)
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestMemoryWritePosMatrixRoundTrip(t *testing.T) {
	m := NewMemory()
	want := PosMatrix{
		{1, 0, 0, 10},
		{0, 1, 0, 20},
		{0, 0, 1, 30},
	}
	m.WritePosMatrix(300, want)
	if got := m.ReadPosMatrix(300); got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestMemoryWriteNormalMatrixRoundTrip(t *testing.T) {
	m := NewMemory()
	want := NormalMatrix{
		{2, 0, 0},
		{0, 3, 0},
		{0, 0, 4},
	}
	m.WriteNormalMatrix(400, want)
	if got := m.ReadNormalMatrix(400); got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
