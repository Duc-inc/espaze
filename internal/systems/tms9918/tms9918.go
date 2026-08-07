// Package tms9918 implements the TI TMS9918(A) video chip from
// scratch: a well-documented, widely-reused classic (ColecoVision,
// SG-1000, MSX1, and several other systems this project's other
// packages build on all share this exact chip). This implementation
// fixes the name/pattern/color table VRAM layout the same way this
// project's other tile-based video chips do, rather than reading the
// real chip's configurable base-address registers, and covers one
// representative graphics mode (closest to real "Graphics Mode II")
// plus sprites - Text Mode and Multicolor Mode aren't implemented.
package tms9918

import "github.com/Duc-inc/espaze/internal/emulation/video"

const (
	Width  = 256
	Height = 192
)

const cyclesPerLine = 228 // approximate Z80-cycle length of one scanline at 3.58MHz
const totalScanlines = 262

// Fixed VRAM layout this project uses instead of real hardware's
// configurable table-base registers.
const (
	nameTableBase     = 0x0000 // 32x24 = 768 bytes
	patternTableBase  = 0x0800 // 256 patterns x 8 bytes = 2KB
	colorTableBase    = 0x2000 // one color byte per name-table tile position (768 bytes)
	spriteAttrBase    = 0x3000 // 32 sprites x 4 bytes
	spritePatternBase = 0x3800 // 256 patterns x 8 bytes
)

// TMS9918 holds 16KB of VRAM and every register.
type TMS9918 struct {
	vram [0x4000]byte

	addr       uint16
	addrLatch  byte
	latched    bool
	writeMode  bool
	readBuffer byte

	regs [8]byte

	status byte

	line      int
	lineClock int

	frame *video.FrameBuffer
}

// New returns a TMS9918 with a blank frame and every register zeroed.
func New() *TMS9918 { return &TMS9918{frame: video.NewFrameBuffer(Width, Height)} }

// Reset clears every register but keeps VRAM and the frame buffer instance.
func (t *TMS9918) Reset() {
	vram, frame := t.vram, t.frame
	*t = TMS9918{vram: vram, frame: frame}
}

// FrameBuffer returns the most recently rendered picture.
func (t *TMS9918) FrameBuffer() *video.FrameBuffer { return t.frame }

func (t *TMS9918) displayEnabled() bool { return t.regs[1]&0x40 != 0 }
func (t *TMS9918) irqEnabled() bool     { return t.regs[1]&0x20 != 0 }
func (t *TMS9918) spritesEnabled() bool { return true }

// IRQVBlank is the bit Step returns when it wants VBlank asserted.
const IRQVBlank = 1 << 0

// Step advances the chip by cpuCycles Z80 cycles.
func (t *TMS9918) Step(cpuCycles int) byte {
	var irq byte
	t.lineClock += cpuCycles
	for t.lineClock >= cyclesPerLine {
		t.lineClock -= cyclesPerLine
		irq |= t.advanceLine()
	}
	return irq
}

func (t *TMS9918) advanceLine() byte {
	var irq byte
	if t.line < Height {
		t.renderScanline(t.line)
	} else if t.line == Height {
		t.status |= 0x80
		if t.irqEnabled() {
			irq |= IRQVBlank
		}
	}
	t.line++
	if t.line >= totalScanlines {
		t.line = 0
	}
	return irq
}
