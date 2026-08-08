// Package adpcm decodes GameCube DSP-ADPCM: the compressed audio
// format real GameCube games ship instead of raw PCM, which
// internal/systems/gamecube/audio's mixer previously assumed it would
// always receive. It's a predictive 4-bit codec: each 8-byte frame
// carries a 1-byte header (a predictor-coefficient index and a scale
// shift) plus 14 4-bit samples, each reconstructed from the running
// prediction history and a per-sound coefficient table that
// accompanies the compressed data itself (not something this project
// derives).
//
// Confidence note: this is a widely documented, standard algorithm
// across public GameCube homebrew/audio-tooling references (needed by
// any tool that reads GC audio at all, not specific to any emulator),
// and this project's implementation follows that common description
// with reasonable confidence - but it hasn't been independently
// re-verified against an authoritative specification this session, so
// treat exact behavior (in particular the >>11 prediction scaling) as
// this project's best understanding rather than a guaranteed-bit-
// exact match to real hardware.
package adpcm

// Coefficients is the 16-pair predictor coefficient table a GC-ADPCM
// stream's header byte indexes into (4 bits -> 16 possible entries).
type Coefficients [16][2]int16

// Decoder holds the running two-sample prediction history GC-ADPCM's
// predictive codec carries between frames.
type Decoder struct {
	coefs        Coefficients
	hist1, hist2 int16
}

// NewDecoder returns a Decoder starting from silence (zeroed history,
// matching a real stream's initial state).
func NewDecoder(coefs Coefficients) *Decoder {
	return &Decoder{coefs: coefs}
}

// DecodeFrame decodes one 8-byte ADPCM frame (1 header byte + 7 data
// bytes, 14 packed 4-bit samples) into 14 PCM samples, continuing
// from this Decoder's running history.
func (d *Decoder) DecodeFrame(frame [8]byte) [14]int16 {
	header := frame[0]
	coef1 := int32(d.coefs[header>>4][0])
	coef2 := int32(d.coefs[header>>4][1])
	scale := uint(header & 0x0F)

	var out [14]int16
	for i := 0; i < 14; i++ {
		var nibble byte
		if i%2 == 0 {
			nibble = frame[1+i/2] >> 4
		} else {
			nibble = frame[1+i/2] & 0x0F
		}
		signed := int32(int8(nibble<<4) >> 4) // sign-extend the 4-bit value

		predicted := (coef1*int32(d.hist1) + coef2*int32(d.hist2)) >> 11
		sample := clamp16(predicted + signed<<scale)

		d.hist2 = d.hist1
		d.hist1 = sample
		out[i] = sample
	}
	return out
}

func clamp16(v int32) int16 {
	switch {
	case v > 32767:
		return 32767
	case v < -32768:
		return -32768
	default:
		return int16(v)
	}
}
