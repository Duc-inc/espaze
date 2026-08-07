// Package tia implements the Atari 2600's Television Interface
// Adapter from scratch: unlike every other video chip in this
// project, real TIA hardware has no framebuffer or sprite table at
// all - games "race the beam", writing registers at precise moments
// as the CRT's electron beam sweeps across the screen to place
// graphics. This implementation reproduces that by stepping one color
// clock at a time rather than rendering whole scanlines/frames in one
// shot, so a game's exact WSYNC/RESP0/HMOVE timing still produces the
// right picture.
package tia

import "github.com/Duc-inc/espaze/internal/emulation/video"

// Width/Height are the TIA's standard visible picture; Height assumes
// the common 192-line/40-line-vblank NTSC timing nearly every
// commercial cartridge targets, rather than tracking the real
// VSYNC/VBLANK register timing exactly.
const (
	Width  = 160
	Height = 192
)

const colorClocksPerLine = 228
const hblankClocks = 68
const totalScanlines = 262
const firstVisibleLine = 40

// TIA holds every register and object needed to reproduce the visible
// picture and simplified 2-channel audio.
type TIA struct {
	vsync, vblank bool

	pf     playfield
	colupf byte
	bg     byte // COLUBK
	p0     player
	p1     player
	m0     movable
	m1     movable
	bl     movable
	a0     audioChannel
	a1     audioChannel

	hmp0, hmp1, hmm0, hmm1, hmbl byte

	cxm0p, cxm1p, cxp0fb, cxp1fb, cxm0fb, cxm1fb, cxblpf, cxppmm byte

	inputLatches [6]byte // INPT0-5, digital-only (paddles read as 0)

	clock, line int
	wsync       bool
	frameDone   bool

	sampleCycles float64
	samples      []int16

	frame *video.FrameBuffer
}

// New returns a TIA with a blank frame and every register zeroed.
func New() *TIA {
	return &TIA{p0: newPlayer(), p1: newPlayer(), m0: newMovable(), m1: newMovable(), bl: newMovable(),
		a0: newAudioChannel(), a1: newAudioChannel(), frame: video.NewFrameBuffer(Width, Height)}
}

// Reset clears every register but keeps the frame buffer instance.
func (t *TIA) Reset() {
	frame := t.frame
	*t = TIA{p0: newPlayer(), p1: newPlayer(), m0: newMovable(), m1: newMovable(), bl: newMovable(),
		a0: newAudioChannel(), a1: newAudioChannel(), frame: frame}
}

// FrameBuffer returns the most recently rendered picture.
func (t *TIA) FrameBuffer() *video.FrameBuffer { return t.frame }

// WSyncPending reports whether the CPU should be halted until the
// start of the next scanline (a WSYNC strobe was just written).
func (t *TIA) WSyncPending() bool { return t.wsync }

// FrameDone reports (and clears) whether a full 262-line raster just
// completed.
func (t *TIA) FrameDone() bool {
	done := t.frameDone
	t.frameDone = false
	return done
}

// Step advances the TIA by cpuCycles CPU cycles (3 color clocks each,
// the fixed ratio on real hardware), rendering pixels and generating
// audio samples along the way.
func (t *TIA) Step(cpuCycles int) {
	for i := 0; i < cpuCycles; i++ {
		t.a0.tick()
		t.a1.tick()
		t.tickSample()

		for c := 0; c < 3; c++ {
			t.tickColorClock()
		}
	}
}

func (t *TIA) tickColorClock() {
	if t.clock >= hblankClocks && t.line >= firstVisibleLine && t.line < firstVisibleLine+Height {
		t.renderPixel(t.clock-hblankClocks, t.line-firstVisibleLine)
	}

	t.clock++
	if t.clock >= colorClocksPerLine {
		t.clock = 0
		t.wsync = false
		t.line++
		if t.line >= totalScanlines {
			t.line = 0
			t.frameDone = true
		}
	}
}

const audioSampleRate = 44100.0
const cpuClockHz = 1193182.0
const cyclesPerSample = cpuClockHz / audioSampleRate

func (t *TIA) tickSample() {
	t.sampleCycles++
	if t.sampleCycles >= cyclesPerSample {
		t.sampleCycles -= cyclesPerSample
		sum := int32(t.a0.sample()) + int32(t.a1.sample())
		t.samples = append(t.samples, clampSample(sum))
	}
}

func clampSample(v int32) int16 {
	switch {
	case v > 32767:
		return 32767
	case v < -32768:
		return -32768
	default:
		return int16(v)
	}
}

// DrainSamples returns and clears every sample generated since the last call.
func (t *TIA) DrainSamples() []int16 {
	out := t.samples
	t.samples = nil
	return out
}
