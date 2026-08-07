package vdp

import "github.com/Duc-inc/espaze/internal/emulation/video"

// Screen dimensions of the SMS's standard (192-line) display mode -
// the one the overwhelming majority of the library uses; the rarer
// 224/240-line Mode 4 variants aren't implemented.
const (
	Width  = 256
	Height = 192
)

// cyclesPerLine is the well-documented SMS constant: one scanline takes
// exactly 228 Z80 T-states.
const cyclesPerLine = 228
const totalScanlines = 262 // NTSC

// VDP is a from-scratch implementation of the SMS's 315-5124-family
// video chip in Mode 4: a scrolling background (with per-tile
// palette/flip/priority) plus up to 64 sprites (8 per line, no flip, no
// priority bit - much simpler than the background layer), rendered one
// scanline at a time like this project's other PPU-family cores.
type VDP struct {
	vram [0x4000]byte
	cram cram
	regs [11]byte

	addr        uint16
	ctrlLow     byte
	ctrlLatched bool
	mode        accessMode
	readBuffer  byte

	status byte

	lineCounter    byte
	lineIRQPending bool

	lineClock int
	line      int

	frame *video.FrameBuffer
}

// New returns a VDP with a blank frame and every register zeroed,
// matching the state a cartridge's init code expects to find at boot.
func New() *VDP {
	return &VDP{frame: video.NewFrameBuffer(Width, Height)}
}

// Reset clears all VDP state except VRAM/CRAM (LoadROM callers clear
// those separately if they need to) and the frame buffer instance.
func (v *VDP) Reset() {
	vram, cr, frame := v.vram, v.cram, v.frame
	*v = VDP{vram: vram, cram: cr, frame: frame}
}

// FrameBuffer returns the most recently rendered picture.
func (v *VDP) FrameBuffer() *video.FrameBuffer { return v.frame }

// CurrentLine returns the scanline the VDP is currently on, i.e. the V
// counter port's value - a few games poll this for raster timing.
// Real hardware's V counter has a well-known non-linear jump in the
// blanking region that this simplifies away to the raw line number.
func (v *VDP) CurrentLine() int { return v.line }

// decodeTilePixel reads one pixel's 4-bit color index out of an 8x8
// Mode 4 tile: 32 bytes per tile, 4 bytes (one per bit-plane) per row,
// most-significant-bit-first within each plane byte.
func (v *VDP) decodeTilePixel(tileAddr uint16, x, y int) byte {
	addr := tileAddr + uint16(y)*4
	b0 := v.vram[addr&0x3FFF]
	b1 := v.vram[(addr+1)&0x3FFF]
	b2 := v.vram[(addr+2)&0x3FFF]
	b3 := v.vram[(addr+3)&0x3FFF]
	bit := 7 - x
	return ((b3>>bit)&1)<<3 | ((b2>>bit)&1)<<2 | ((b1>>bit)&1)<<1 | (b0>>bit)&1
}

func (v *VDP) setPixel(x, y int, r, g, b byte) {
	v.frame.SetPixel(x, y, r, g, b, 0xFF)
}

// Step advances the VDP by cpuCycles T-states and reports which
// interrupt lines it wants asserted this call: bit0 for the frame
// (VBlank) interrupt, bit1 for the line interrupt.
const (
	IRQFrame = 1 << 0
	IRQLine  = 1 << 1
)

func (v *VDP) Step(cpuCycles int) byte {
	var irq byte
	v.lineClock += cpuCycles

	for v.lineClock >= cyclesPerLine {
		v.lineClock -= cyclesPerLine
		irq |= v.advanceLine()
	}
	return irq
}

func (v *VDP) advanceLine() byte {
	var irq byte

	if v.line <= Height {
		if v.lineCounter == 0 {
			v.lineCounter = v.regs[10]
			if v.lineIRQEnabled() {
				v.lineIRQPending = true
			}
		} else {
			v.lineCounter--
		}
	} else {
		v.lineCounter = v.regs[10]
	}
	if v.lineIRQPending {
		irq |= IRQLine
	}

	if v.line < Height {
		v.renderScanline(v.line)
	} else if v.line == Height {
		v.status |= statusVBlank
		if v.frameIRQEnabled() {
			irq |= IRQFrame
		}
	}

	v.line++
	if v.line >= totalScanlines {
		v.line = 0
	}
	return irq
}

func (v *VDP) renderScanline(line int) {
	if !v.displayEnabled() {
		for x := 0; x < Width; x++ {
			v.setPixel(x, line, 0, 0, 0)
		}
		return
	}

	var bg [Width]bgPixel
	v.renderBackgroundLine(line, &bg)
	if v.regs[0]&reg0MaskColumn0 != 0 {
		r, g, b := v.cram.rgb(v.backdropColor())
		for x := 0; x < 8; x++ {
			v.setPixel(x, line, r, g, b)
		}
	}
	v.renderSpritesLine(line, &bg)
}
