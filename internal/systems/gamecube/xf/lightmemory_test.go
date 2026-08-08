package xf

import "testing"

func TestLightIndexForAddr(t *testing.T) {
	if idx, ok := LightIndexForAddr(LightsStart); !ok || idx != 0 {
		t.Fatalf("LightIndexForAddr(LightsStart) = (%d,%v), want (0,true)", idx, ok)
	}
	if idx, ok := LightIndexForAddr(LightsStart + LightWords + 2); !ok || idx != 1 {
		t.Fatalf("light 1 address = (%d,%v), want (1,true)", idx, ok)
	}
	if _, ok := LightIndexForAddr(LightsStart - 1); ok {
		t.Fatal("expected addresses before LightsStart to not resolve to a light")
	}
	if _, ok := LightIndexForAddr(LightsEnd); ok {
		t.Fatal("expected LightsEnd itself to not resolve to a light")
	}
}

func TestReadLightDecodesPositionAndColor(t *testing.T) {
	mem := NewMemory()
	base := uint16(LightsStart + 2*LightWords) // light 2
	mem.WriteFloat32(base+LightPositionOffset, 1)
	mem.WriteFloat32(base+LightPositionOffset+1, 2)
	mem.WriteFloat32(base+LightPositionOffset+2, 3)
	mem.Write(base+LightColorOffset, uint32(0xFF804000)) // R=255,G=128,B=64,A=0

	l := mem.ReadLight(2)
	if l.Position != (Vec3{X: 1, Y: 2, Z: 3}) {
		t.Fatalf("Position = %+v, want (1,2,3)", l.Position)
	}
	want := LightColor{R: 1, G: 128.0 / 255, B: 64.0 / 255}
	if l.Color != want {
		t.Fatalf("Color = %+v, want %+v", l.Color, want)
	}
	if !l.Enabled {
		t.Fatal("expected Enabled to be true")
	}
}

func TestReadLightZeroForUntouchedLight(t *testing.T) {
	mem := NewMemory()
	l := mem.ReadLight(5)
	if l.Position != (Vec3{}) || l.Color != (LightColor{}) {
		t.Fatalf("got %+v, want zero position/color", l)
	}
}
