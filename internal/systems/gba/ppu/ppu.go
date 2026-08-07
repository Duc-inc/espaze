// Package ppu implements the Game Boy Advance's PPU from scratch,
// covering Mode 0 (4 tiled, non-affine backgrounds) and Mode 3 (a
// single 15-bit direct-color bitmap) plus regular (non-affine)
// sprites - the two video modes and sprite behavior that cover the
// large majority of simple 2D GBA software. Modes 1/2 (mixed
// tiled/affine) and 4/5 (paletted/smaller bitmaps), affine
// backgrounds and sprites, mosaic, and the alpha-blend/window special
// effects aren't implemented - a deliberate scope decision consistent
// with this project's other video chips, not an oversight.
package ppu

import "github.com/Duc-inc/espaze/internal/emulation/video"

const (
	Width  = 240
	Height = 160
)

const cyclesPerDot = 4
const dotsPerLine = 308
const totalScanlines = 228

// PPU holds VRAM, OAM, palette RAM, and every register this project implements.
type PPU struct {
	vram    [0x18000]byte // 96KB
	oam     [0x400]byte   // 1KB, 128 sprites x 8 bytes
	palette [512]uint16   // 256 BG + 256 OBJ, RGB555

	dispcnt  uint16
	dispstat uint16
	bgcnt    [4]uint16
	bghofs   [4]uint16
	bgvofs   [4]uint16

	dot, line int

	frame *video.FrameBuffer
}

// New returns a PPU with a blank frame and every register zeroed.
func New() *PPU { return &PPU{frame: video.NewFrameBuffer(Width, Height)} }

// Reset clears every register but keeps VRAM/OAM/palette and the
// frame buffer instance - matching this project's other video chips.
func (p *PPU) Reset() {
	vram, oam, pal, frame := p.vram, p.oam, p.palette, p.frame
	*p = PPU{vram: vram, oam: oam, palette: pal, frame: frame}
}

// FrameBuffer returns the most recently rendered picture.
func (p *PPU) FrameBuffer() *video.FrameBuffer { return p.frame }

// IRQVBlank is the bit Step returns when it wants VBlank asserted -
// HBlank and the raster-compare (VCOUNT) interrupts aren't implemented.
const IRQVBlank = 1 << 0

func (p *PPU) mode() int { return int(p.dispcnt & 0x07) }

func (p *PPU) bgEnabled(n int) bool           { return p.dispcnt&(1<<uint(8+n)) != 0 }
func (p *PPU) objEnabled() bool               { return p.dispcnt&0x1000 != 0 }
func (p *PPU) vblankIRQEnabled(v uint16) bool { return v&0x08 != 0 }

// Step advances the PPU by cpuCycles CPU cycles.
func (p *PPU) Step(cpuCycles int) byte {
	var irq byte
	p.dot += cpuCycles
	for p.dot >= dotsPerLine*cyclesPerDot {
		p.dot -= dotsPerLine * cyclesPerDot
		irq |= p.advanceLine()
	}
	return irq
}

// WriteDISPSTAT implements the display status register's writable bits.
func (p *PPU) WriteDISPSTAT(v uint16) { p.dispstat = v }

func (p *PPU) advanceLine() byte {
	var irq byte
	if p.line < Height {
		p.renderScanline(p.line)
	} else if p.line == Height {
		if p.vblankIRQEnabled(p.dispstat) {
			irq |= IRQVBlank
		}
	}
	p.line++
	if p.line >= totalScanlines {
		p.line = 0
	}
	return irq
}
