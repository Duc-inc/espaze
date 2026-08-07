// Package vdc implements the PC Engine's HuC6270 Video Display
// Controller: its own 64KB (32K-word) VRAM, a scrolling tile
// background, and up to 64 sprites, resolved to final RGB through a
// PaletteResolver (the VCE, see the vce package) exactly like real
// hardware splits color resolution into a separate chip. This project
// fixes the picture at a single common resolution (256x224) and a
// single background/sprite tile layout (see background.go/sprites.go)
// rather than reproducing the real chip's several configurable screen
// modes and arbitrary VRAM placement - a deliberate simplification in
// the same spirit as this project's other video chips.
package vdc

import "github.com/Duc-inc/espaze/internal/emulation/video"

const (
	Width  = 256
	Height = 224
)

const cyclesPerLine = 455 // approximate HuC6280-cycle length of one scanline at the common 7.16MHz rate
const totalScanlines = 262

// PaletteResolver turns a 9-bit VCE color index into RGB - implemented
// by the vce package.
type PaletteResolver interface {
	Resolve(index uint16) (r, g, b byte)
}

// VDC holds VRAM, the sprite attribute table, and every register.
type VDC struct {
	vram [0x8000]uint16
	sat  [64 * 4]uint16

	selectedReg  byte
	regs         [20]uint16
	writeHiNext  bool // most VDC registers latch low byte first, then high
	vramLowLatch byte

	palette PaletteResolver

	line      int
	lineClock int

	frame *video.FrameBuffer
}

// New wires a VDC to the palette resolver it renders through.
func New(palette PaletteResolver) *VDC {
	return &VDC{palette: palette, frame: video.NewFrameBuffer(Width, Height)}
}

// Reset clears every register but keeps VRAM/SAT/the frame buffer instance.
func (v *VDC) Reset() {
	vram, sat, frame, pal := v.vram, v.sat, v.frame, v.palette
	*v = VDC{vram: vram, sat: sat, frame: frame, palette: pal}
}

// FrameBuffer returns the most recently rendered picture.
func (v *VDC) FrameBuffer() *video.FrameBuffer { return v.frame }

// IRQVBlank/IRQLine are the bits Step returns when it wants an
// interrupt asserted (both route to the CPU's IRQ1 line on real
// hardware).
const (
	IRQVBlank = 1 << 0
	IRQLine   = 1 << 1
)

func (v *VDC) vblankIRQEnabled() bool { return v.regs[5]&0x08 != 0 }
func (v *VDC) lineIRQEnabled() bool   { return v.regs[5]&0x04 != 0 }
func (v *VDC) bgEnabled() bool        { return v.regs[5]&0x80 != 0 }
func (v *VDC) spritesEnabled() bool   { return v.regs[5]&0x40 != 0 }

// Step advances the VDC by cpuCycles HuC6280 cycles.
func (v *VDC) Step(cpuCycles int) byte {
	var irq byte
	v.lineClock += cpuCycles
	for v.lineClock >= cyclesPerLine {
		v.lineClock -= cyclesPerLine
		irq |= v.advanceLine()
	}
	return irq
}

func (v *VDC) advanceLine() byte {
	var irq byte

	if v.line < Height {
		v.renderScanline(v.line)
	} else if v.line == Height {
		if v.vblankIRQEnabled() {
			irq |= IRQVBlank
		}
	}

	if int(v.regs[6]) == v.line && v.lineIRQEnabled() {
		irq |= IRQLine
	}

	v.line++
	if v.line >= totalScanlines {
		v.line = 0
	}
	return irq
}
