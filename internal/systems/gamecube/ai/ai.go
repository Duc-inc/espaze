// Package ai implements the GameCube's Audio Streaming Interface: the
// real memory-mapped registers that start/stop audio streaming,
// control volume, and count output samples - wiring the existing
// audio.Mixer into something real game code actually pokes, instead
// of only Go-level Set calls. Addresses and bit layout come from a
// public hardware register reference (YAGCD chapter 5, "AI - Audio
// Streaming Interface").
package ai

const (
	Base = 0xCC006C00
	Size = 0x20

	regAICR   = 0x00
	regAIVR   = 0x04
	regAISCNT = 0x08
	regAIIT   = 0x0C

	bitPSTAT = 1 << 0 // playing status: streaming clock enabled
	bitAIINT = 1 << 2 // interrupt status/clear
)

// AI holds the Audio Interface's register state.
type AI struct {
	playing     bool
	interrupt   bool
	volumeL     byte
	volumeR     byte
	sampleCnt   uint32
	interruptAt uint32
}

func New() *AI { return &AI{} }

// Playing reports whether streaming audio is enabled (AICR's PSTAT
// bit) - what should gate audio.Mixer.Step actually advancing.
func (a *AI) Playing() bool { return a.playing }

// Volume returns the left/right channel volumes (0-255).
func (a *AI) Volume() (l, r byte) { return a.volumeL, a.volumeR }

// Interrupting reports whether AI's sample-count interrupt is
// currently active (not yet cleared by a game writing AICR's AIINT
// bit) - the level-triggered cause signal pi.PI's AI bit reports (see
// gamecube.go's Step).
func (a *AI) Interrupting() bool { return a.interrupt }

// Step advances the sample counter by one stereo sample while playing
// and reports whether the interrupt-timing match just fired.
func (a *AI) Step() (interrupted bool) {
	if !a.playing {
		return false
	}
	a.sampleCnt++
	if a.sampleCnt == a.interruptAt {
		a.interrupt = true
		interrupted = true
	}
	return interrupted
}

func (a *AI) Read32(offset uint32) uint32 {
	switch offset {
	case regAICR:
		var v uint32
		if a.playing {
			v |= bitPSTAT
		}
		if a.interrupt {
			v |= bitAIINT
		}
		return v
	case regAIVR:
		return uint32(a.volumeR)<<8 | uint32(a.volumeL)
	case regAISCNT:
		return a.sampleCnt
	case regAIIT:
		return a.interruptAt
	default:
		return 0
	}
}

func (a *AI) Write32(offset uint32, val uint32) {
	switch offset {
	case regAICR:
		a.playing = val&bitPSTAT != 0
		if val&bitAIINT != 0 {
			a.interrupt = false // write-1-to-clear
		}
	case regAIVR:
		a.volumeL = byte(val)
		a.volumeR = byte(val >> 8)
	case regAIIT:
		a.interruptAt = val
	}
}
