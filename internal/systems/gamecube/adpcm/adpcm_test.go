package adpcm

import "testing"

func TestDecodeFrameWithZeroCoefficientsPassesNibblesThrough(t *testing.T) {
	// coef1=coef2=0 -> prediction is always 0, so with scale=0 each
	// sample is exactly its signed 4-bit nibble value.
	var coefs Coefficients // all zero
	d := NewDecoder(coefs)

	frame := [8]byte{
		0x00,       // header: predictor 0, scale 0
		0x1F,       // nibbles: 1, -1
		0x87,       // nibbles: -8, 7
		0, 0, 0, 0, // remaining nibbles: 0
	}
	out := d.DecodeFrame(frame)

	want := [14]int16{1, -1, -8, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if out != want {
		t.Fatalf("got %v, want %v", out, want)
	}
}

func TestDecodeFrameAppliesScaleShift(t *testing.T) {
	var coefs Coefficients
	d := NewDecoder(coefs)

	frame := [8]byte{
		0x03, // header: predictor 0, scale 3
		0x10, // first nibble: 1 -> 1<<3 = 8
		0, 0, 0, 0, 0, 0,
	}
	out := d.DecodeFrame(frame)
	if out[0] != 8 {
		t.Fatalf("out[0] = %d, want 8", out[0])
	}
}

func TestDecodeFramePredictsFromHistoryAcrossFrames(t *testing.T) {
	var coefs Coefficients
	coefs[1] = [2]int16{2048, 0} // predictor 1: coef1=2048 (=1.0 in Q11), coef2=0
	d := NewDecoder(coefs)

	// Frame 1 (predictor 0, zero coefficients, scale 7): every nibble
	// is 1, so every sample is 1<<7 = 128 - including the last one,
	// which becomes this decoder's carried-forward history.
	frame1 := [8]byte{
		0x07,                                           // predictor 0, scale 7
		0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, // 14 nibbles, all 1
	}
	out1 := d.DecodeFrame(frame1)
	if out1[13] != 128 {
		t.Fatalf("out1[13] = %d, want 128", out1[13])
	}

	// Frame 2 (predictor 1: coef1=2048, coef2=0, scale 0, nibble 0):
	// predicted = (2048*hist1 + 0*hist2) >> 11 = hist1 exactly, since
	// hist1 is 128 (a multiple of 1, dividing exactly).
	frame2 := [8]byte{
		0x10, // predictor 1, scale 0
		0x00, // first nibble: 0
		0, 0, 0, 0, 0, 0,
	}
	out2 := d.DecodeFrame(frame2)
	if out2[0] != 128 {
		t.Fatalf("out2[0] = %d, want 128 (predicted forward from frame 1's last sample)", out2[0])
	}
}

func TestDecodeFrameClampsOverflow(t *testing.T) {
	var coefs Coefficients
	coefs[0] = [2]int16{32767, 0}
	d := NewDecoder(coefs)
	d.hist1 = 32767 // seed a large history directly to force overflow

	frame := [8]byte{
		0x00, // predictor 0, scale 0
		0x00, // nibble 0: no extra contribution, prediction alone overflows
		0, 0, 0, 0, 0, 0,
	}
	out := d.DecodeFrame(frame)
	if out[0] != 32767 {
		t.Fatalf("out[0] = %d, want clamped to 32767", out[0])
	}
}
