// Package ppu implements a from-scratch, simplified slice of the
// SNES's S-PPU: up to 4 tiled background layers (2bpp or 4bpp,
// configurable per layer - covering the common "Mode 1" shape real
// games use most) plus sprites, resolved through CGRAM's 256-color
// BGR555 palette. Real hardware's Mode 7 (rotation/scaling), the
// higher-color-depth modes, color math (additive/subtractive
// blending), windowing, and mosaic aren't implemented - a deliberate
// scope decision consistent with this project's other video chips,
// not an oversight.
package ppu

import "github.com/Duc-inc/espaze/internal/emulation/video"

const (
	Width  = 256
	Height = 224
)

const cyclesPerLine = 1364 // approximate 65816-cycle length of one scanline (master clock/4)
const totalScanlines = 262

// PPU holds VRAM, OAM, CGRAM, and every register.
type PPU struct {
	vram  [0x10000]uint16 // 64K words
	oam   [0x220]byte     // 128 sprites x 4 bytes + 32 bytes high table
	cgram [256]uint16     // BGR555

	bgcnt   [4]byte // per-layer: bit0-1=bpp select(0=2bpp,1=4bpp), bits2-15=tilemap/char base packed simply
	bgHOfs  [4]uint16
	bgVOfs  [4]uint16
	bgMain  byte // bit n: layer n enabled on main screen
	objMain bool

	vramAddr    uint16
	vramLowNext bool

	cgramAddr byte
	cgramHigh bool

	line, lineClock int

	frame *video.FrameBuffer
}

// New returns a PPU with a blank frame and every register zeroed.
func New() *PPU { return &PPU{frame: video.NewFrameBuffer(Width, Height)} }

// Reset clears every register but keeps VRAM/OAM/CGRAM and the frame
// buffer instance.
func (p *PPU) Reset() {
	vram, oam, cgram, frame := p.vram, p.oam, p.cgram, p.frame
	*p = PPU{vram: vram, oam: oam, cgram: cgram, frame: frame}
}

// FrameBuffer returns the most recently rendered picture.
func (p *PPU) FrameBuffer() *video.FrameBuffer { return p.frame }

// IRQVBlank is the bit Step returns when it wants VBlank/NMI asserted.
const IRQVBlank = 1 << 0

// Step advances the PPU by cpuCycles 65816 cycles.
func (p *PPU) Step(cpuCycles int) byte {
	var irq byte
	p.lineClock += cpuCycles
	for p.lineClock >= cyclesPerLine {
		p.lineClock -= cyclesPerLine
		irq |= p.advanceLine()
	}
	return irq
}

func (p *PPU) advanceLine() byte {
	var irq byte
	if p.line < Height {
		p.renderScanline(p.line)
	} else if p.line == Height {
		irq |= IRQVBlank
	}
	p.line++
	if p.line >= totalScanlines {
		p.line = 0
	}
	return irq
}
