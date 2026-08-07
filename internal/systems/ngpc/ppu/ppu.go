// Package ppu implements a from-scratch, simplified video chip for
// the Neo Geo Pocket Color: a single scrollable tiled background plus
// sprites, resolved through a 4096-color (12-bit RGB) palette - the
// real hardware's actual chip is even less publicly documented than
// the TLCS900H CPU itself (see the cpu package's doc comment for that
// caveat), including real detail like its two independently
// scrollable background planes and their blend/priority rules; this
// project implements one plane instead of two, a further,
// self-acknowledged simplification on top of the CPU's own.
package ppu

import "github.com/Duc-inc/espaze/internal/emulation/video"

const (
	Width  = 160
	Height = 152
)

const cyclesPerLine = 515 // approximate TLCS900H-cycle length of one scanline
const totalScanlines = 198

// PPU holds VRAM, sprite table, palette RAM, and every register.
type PPU struct {
	vram    [0x4000]byte // 16KB: name table + tile patterns
	sprites [64 * 4]byte // 64 sprites x 4 bytes (Y, X, tile, attr)
	palette [32]uint16   // 2 banks x 16 colors, 12-bit RGB (-RRRRGGGGBBBB)

	scrollX, scrollY    byte
	bgEnable, objEnable bool

	line, lineClock int

	frame *video.FrameBuffer
}

// New returns a PPU with a blank frame and every register zeroed.
func New() *PPU { return &PPU{frame: video.NewFrameBuffer(Width, Height)} }

// Reset clears every register but keeps VRAM/sprites/palette and the
// frame buffer instance.
func (p *PPU) Reset() {
	vram, sprites, pal, frame := p.vram, p.sprites, p.palette, p.frame
	*p = PPU{vram: vram, sprites: sprites, palette: pal, frame: frame}
}

// FrameBuffer returns the most recently rendered picture.
func (p *PPU) FrameBuffer() *video.FrameBuffer { return p.frame }

// IRQVBlank is the bit Step returns when it wants VBlank asserted.
const IRQVBlank = 1 << 0

// Step advances the PPU by cpuCycles CPU cycles.
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
