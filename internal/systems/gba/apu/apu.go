// Package apu implements a simplified slice of the Game Boy Advance's
// audio hardware: the two Direct Sound FIFO channels (A/B), which
// carry the large majority of later commercial GBA games' music and
// sound effects as streamed PCM. The four legacy Game Boy-style
// channels (square/square/wave/noise) real hardware also has aren't
// implemented - a real gap for games that lean on them for
// chiptune-style audio, documented rather than silently dropped.
package apu

// SampleRate is the PCM rate every generated sample chunk is at.
const SampleRate = 44100

const cpuClockHz = 16777216.0

// fifoDrainHz is the rate this project drains each Direct Sound FIFO
// at - real hardware's rate is set by whichever timer a game
// configures to refill it (commonly ~18-32kHz); fixing it here
// instead of wiring the timer module through is a deliberate
// simplification.
const fifoDrainHz = 32768.0
const cyclesPerDrain = cpuClockHz / fifoDrainHz
const cyclesPerSample = cpuClockHz / SampleRate

// APU holds both Direct Sound FIFOs and their mix settings.
type APU struct {
	fifoA, fifoB []int8
	lastA, lastB int8

	volA, volB       bool // false=50%, true=100%
	enableA, enableB bool

	drainCycles  float64
	sampleCycles float64
	samples      []int16
}

// New returns an APU with both FIFOs empty.
func New() *APU { return &APU{} }

// Reset empties both FIFOs and clears mix settings.
func (a *APU) Reset() { *a = APU{} }

// WriteFIFOA/WriteFIFOB push one byte (a signed 8-bit PCM sample) into
// the given channel's FIFO - real hardware queues up to 32 bytes;
// this project doesn't cap the queue, relying on the caller (DMA)
// pacing writes reasonably.
func (a *APU) WriteFIFOA(v byte) { a.fifoA = append(a.fifoA, int8(v)) }
func (a *APU) WriteFIFOB(v byte) { a.fifoB = append(a.fifoB, int8(v)) }

// WriteSoundCntH implements SOUNDCNT_H's Direct Sound control bits.
func (a *APU) WriteSoundCntH(v uint16) {
	a.volA = v&0x04 != 0
	a.volB = v&0x08 != 0
	a.enableA = v&0x0300 != 0
	a.enableB = v&0x3000 != 0
	if v&0x0800 != 0 {
		a.fifoA = nil
	}
	if v&0x8000 != 0 {
		a.fifoB = nil
	}
}

// Step advances the APU by cpuCycles CPU cycles, draining each FIFO
// at its fixed rate and generating output samples.
func (a *APU) Step(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		a.drainCycles++
		if a.drainCycles >= cyclesPerDrain {
			a.drainCycles -= cyclesPerDrain
			a.drainFIFOs()
		}
		a.sampleCycles++
		if a.sampleCycles >= cyclesPerSample {
			a.sampleCycles -= cyclesPerSample
			a.samples = append(a.samples, a.mixSample())
		}
	}
}

func (a *APU) drainFIFOs() {
	if len(a.fifoA) > 0 {
		a.lastA = a.fifoA[0]
		a.fifoA = a.fifoA[1:]
	}
	if len(a.fifoB) > 0 {
		a.lastB = a.fifoB[0]
		a.fifoB = a.fifoB[1:]
	}
}

func (a *APU) mixSample() int16 {
	var sum int32
	if a.enableA {
		sum += channelSample(a.lastA, a.volA)
	}
	if a.enableB {
		sum += channelSample(a.lastB, a.volB)
	}
	switch {
	case sum > 32767:
		return 32767
	case sum < -32768:
		return -32768
	default:
		return int16(sum)
	}
}

func channelSample(v int8, fullVolume bool) int32 {
	sample := int32(v) * 200
	if !fullVolume {
		sample /= 2
	}
	return sample
}

// DrainSamples returns and clears every sample generated since the last call.
func (a *APU) DrainSamples() []int16 {
	out := a.samples
	a.samples = nil
	return out
}
