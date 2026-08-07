package ym2612

import "math"

// envelope phases, in the order a note actually moves through them.
const (
	envAttack = iota
	envDecay1
	envDecay2
	envRelease
	envOff
)

// operator is one of a channel's 4 sine-wave generators: a phase
// accumulator (frequency comes from the channel, scaled by this
// operator's own multiplier/detune) run through a simplified ADSR
// envelope. Real YM2612 envelopes follow logarithmic rate tables this
// approximates with linear ramps - close enough to produce a
// recognizable FM voice, not a bit-exact reproduction of the chip's
// timbre.
type operator struct {
	mul byte // frequency multiplier, 0 means x0.5
	det byte // detune (0-7, a small +/- shift; only the sign/magnitude bucket is modeled)
	tl  byte // total level: 0 = loudest, 127 = silent
	ar  byte // attack rate (0-31)
	d1r byte // decay rate to sustain level
	d2r byte // decay rate after sustain (a slow further decay)
	sl  byte // sustain level (0-15, 15 = quietest sustain point)
	rr  byte // release rate (0-15)

	phase    float64
	envPhase int
	envLevel float64 // 0 (silent) to 1 (full)
	keyOn    bool
}

func (op *operator) writeMul(v byte) {
	op.det = (v >> 4) & 0x07
	op.mul = v & 0x0F
}
func (op *operator) writeTL(v byte)  { op.tl = v & 0x7F }
func (op *operator) writeAR(v byte)  { op.ar = v & 0x1F }
func (op *operator) writeD1R(v byte) { op.d1r = v & 0x1F }
func (op *operator) writeD2R(v byte) { op.d2r = v & 0x1F }
func (op *operator) writeSLRR(v byte) {
	op.sl = (v >> 4) & 0x0F
	op.rr = v & 0x0F
}

func (op *operator) keyOnEvent() {
	op.keyOn = true
	op.envPhase = envAttack
	op.phase = 0
}

func (op *operator) keyOffEvent() {
	op.keyOn = false
	op.envPhase = envRelease
}

// frequencyRatio turns MUL/DET into a multiplier against the channel's
// base frequency - a coarse approximation of the real detune table.
func (op *operator) frequencyRatio() float64 {
	mul := float64(op.mul)
	if mul == 0 {
		mul = 0.5
	}
	detune := float64(op.det) * 0.006 // small fractional shift per detune step
	return mul + detune
}

const sampleRate = 44100.0

// step advances the phase accumulator and envelope by one sample and
// returns this operator's current output, optionally phase-modulated
// by an upstream operator's output (modIn, in radians).
func (op *operator) step(baseFreq, modIn float64) float64 {
	freq := baseFreq * op.frequencyRatio()
	op.phase += freq / sampleRate
	if op.phase >= 1 {
		op.phase -= math.Floor(op.phase)
	}

	op.tickEnvelope()

	tlAtten := 1.0 - float64(op.tl)/127.0
	return math.Sin(2*math.Pi*op.phase+modIn) * op.envLevel * tlAtten
}

func (op *operator) tickEnvelope() {
	sustainLevel := 1.0 - float64(op.sl)/15.0
	if op.sl == 15 {
		sustainLevel = 0
	}

	switch op.envPhase {
	case envAttack:
		rate := rateToStep(op.ar)
		op.envLevel += rate
		if op.envLevel >= 1 {
			op.envLevel = 1
			op.envPhase = envDecay1
		}
	case envDecay1:
		rate := rateToStep(op.d1r)
		op.envLevel -= rate
		if op.envLevel <= sustainLevel {
			op.envLevel = sustainLevel
			op.envPhase = envDecay2
		}
	case envDecay2:
		rate := rateToStep(op.d2r) * 0.25
		op.envLevel -= rate
		if op.envLevel < 0 {
			op.envLevel = 0
		}
	case envRelease:
		rate := rateToStep(op.rr) * 2
		op.envLevel -= rate
		if op.envLevel <= 0 {
			op.envLevel = 0
			op.envPhase = envOff
		}
	}
}

// rateToStep converts a 0-31 hardware rate into a per-sample envelope
// delta - higher rates move faster, matching the real chip's direction
// even though the exact curve is only approximated.
func rateToStep(rate byte) float64 {
	if rate == 0 {
		return 0
	}
	return float64(rate) * float64(rate) / 6000.0
}
