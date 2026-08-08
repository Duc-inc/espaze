package xf

import "testing"

func TestLightColorAddMulScalarMul(t *testing.T) {
	a := LightColor{R: 0.2, G: 0.3, B: 0.4}
	b := LightColor{R: 0.1, G: 0.1, B: 0.1}

	if got, want := a.Add(b), (LightColor{R: 0.3, G: 0.4, B: 0.5}); got != want {
		t.Fatalf("Add: got %+v, want %+v", got, want)
	}
	if got, want := a.MulScalar(2), (LightColor{R: 0.4, G: 0.6, B: 0.8}); got != want {
		t.Fatalf("MulScalar: got %+v, want %+v", got, want)
	}
	if got, want := a.Mul(LightColor{R: 1, G: 0, B: 0.5}), (LightColor{R: 0.2, G: 0, B: 0.2}); got != want {
		t.Fatalf("Mul: got %+v, want %+v", got, want)
	}
}

func TestLightColorClamp01(t *testing.T) {
	got := LightColor{R: -0.5, G: 0.5, B: 2.0}.Clamp01()
	want := LightColor{R: 0, G: 0.5, B: 1}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestIlluminateAmbientOnlyWithNoLights(t *testing.T) {
	ambient := Ambient{Color: LightColor{R: 0.2, G: 0.2, B: 0.2}}
	material := LightColor{R: 1, G: 1, B: 1}

	got := Illuminate(Vec3{}, Vec3{X: 0, Y: 0, Z: 1}, material, ambient, nil)
	want := LightColor{R: 0.2, G: 0.2, B: 0.2}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestIlluminateDiffuseFullyFacingLight(t *testing.T) {
	// A light directly along the normal direction should produce a
	// full N.L = 1 diffuse contribution.
	light := Light{Position: Vec3{X: 0, Y: 0, Z: 10}, Color: LightColor{R: 1, G: 1, B: 1}, Enabled: true}
	ambient := Ambient{} // no ambient term, isolate the diffuse math

	got := Illuminate(Vec3{}, Vec3{X: 0, Y: 0, Z: 1}, LightColor{R: 1, G: 1, B: 1}, ambient, []Light{light})
	want := LightColor{R: 1, G: 1, B: 1}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestIlluminateSurfaceFacingAwayGetsNoDiffuse(t *testing.T) {
	// The light sits behind the surface relative to its normal, so
	// N.L is negative and should clamp to zero contribution.
	light := Light{Position: Vec3{X: 0, Y: 0, Z: -10}, Color: LightColor{R: 1, G: 1, B: 1}, Enabled: true}
	ambient := Ambient{Color: LightColor{R: 0.1, G: 0.1, B: 0.1}}

	got := Illuminate(Vec3{}, Vec3{X: 0, Y: 0, Z: 1}, LightColor{R: 1, G: 1, B: 1}, ambient, []Light{light})
	want := LightColor{R: 0.1, G: 0.1, B: 0.1}
	if got != want {
		t.Fatalf("got %+v, want %+v (ambient only, diffuse clamped away)", got, want)
	}
}

func TestIlluminateDisabledLightIsIgnored(t *testing.T) {
	light := Light{Position: Vec3{X: 0, Y: 0, Z: 10}, Color: LightColor{R: 1, G: 1, B: 1}, Enabled: false}

	got := Illuminate(Vec3{}, Vec3{X: 0, Y: 0, Z: 1}, LightColor{R: 1, G: 1, B: 1}, Ambient{}, []Light{light})
	want := LightColor{}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestIlluminateClampsOvershootFromMultipleLights(t *testing.T) {
	bright := LightColor{R: 1, G: 1, B: 1}
	l1 := Light{Position: Vec3{X: 0, Y: 0, Z: 10}, Color: bright, Enabled: true}
	l2 := Light{Position: Vec3{X: 0, Y: 0, Z: 10}, Color: bright, Enabled: true}

	got := Illuminate(Vec3{}, Vec3{X: 0, Y: 0, Z: 1}, bright, Ambient{}, []Light{l1, l2})
	want := LightColor{R: 1, G: 1, B: 1}
	if got != want {
		t.Fatalf("got %+v, want %+v (clamped to displayable range)", got, want)
	}
}
