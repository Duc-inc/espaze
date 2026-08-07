package vdp

import "github.com/Duc-inc/espaze/internal/emulation/video"

// Screen dimensions: the common 320x224 "H40" mode nearly every
// Genesis game uses; the rarer 256-pixel-wide "H32" mode isn't
// implemented (the frame buffer is always allocated at H40 width - a
// game running in H32 will render into the left 256 pixels of it).
const (
	Width  = 320
	Height = 224
)

// cyclesPerLine/totalScanlines approximate NTSC timing: 262 scanlines
// per frame at roughly 59.92Hz on a ~7.67MHz 68000 clock.
const cyclesPerLine = 488
const totalScanlines = 262

// VDP is a from-scratch implementation of the Genesis's video chip: two
// independently-scrollable background planes (A and B) plus a sprite
// plane with variable-size, linked-list-ordered sprites - substantially
// more capable than the Master System's VDP this project also
// implements. The Window plane and per-line/per-column scroll
// granularity beyond "whole plane" aren't implemented; see
// renderBackgroundLine's own notes.
type VDP struct {
	vram    [0x10000]byte
	palette cram
	vsram   [40]uint16
	regs    [24]byte

	code        byte
	addr        uint16
	ctrlLow     uint16
	ctrlPending bool

	dmaFillArmed bool
	mem          MemoryReader

	status byte

	lineClock int
	line      int

	frame *video.FrameBuffer
}

// New returns a VDP with a blank frame and every register zeroed.
func New() *VDP {
	return &VDP{frame: video.NewFrameBuffer(Width, Height)}
}

// SetMemory wires the 68k-visible memory space DMA reads from - see
// MemoryReader in dma.go for why this can't just be a constructor arg.
func (v *VDP) SetMemory(mem MemoryReader) { v.mem = mem }

// Reset clears all VDP state except VRAM/CRAM/VSRAM and the frame
// buffer instance.
func (v *VDP) Reset() {
	vram, pal, vsram, frame, mem := v.vram, v.palette, v.vsram, v.frame, v.mem
	*v = VDP{vram: vram, palette: pal, vsram: vsram, frame: frame, mem: mem}
}

// FrameBuffer returns the most recently rendered picture.
func (v *VDP) FrameBuffer() *video.FrameBuffer { return v.frame }

// IRQFrame/IRQLine are the bits Step returns when it wants an
// interrupt asserted.
const (
	IRQFrame = 1 << 0
	IRQLine  = 1 << 1
)

// Step advances the VDP by cpuCycles 68000 cycles and reports which
// interrupt lines it wants asserted.
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

	if v.line < Height {
		v.renderScanline(v.line)
	} else if v.line == Height {
		v.status |= statusVBlank
		if v.vblankIRQEnabled() {
			irq |= IRQFrame
		}
	}

	v.line++
	if v.line >= totalScanlines {
		v.line = 0
		v.status &^= statusVBlank
	}
	return irq
}

// renderScanline resolves one line by walking the real hardware's
// fixed priority order top to bottom - sprite-high, plane-A-high,
// plane-B-high, sprite-normal, plane-A-normal, plane-B-normal,
// backdrop - taking the first layer that's actually opaque at each
// pixel, rather than the simpler (but not hardware-accurate) "sprites
// always win" rule.
func (v *VDP) renderScanline(line int) {
	if !v.displayEnabled() {
		for x := 0; x < Width; x++ {
			v.frame.SetPixel(x, line, 0, 0, 0, 0xFF)
		}
		return
	}

	var planeAIdx, planeBIdx [Width]byte
	var planeAPri, planeBPri [Width]bool
	v.renderPlaneLine(line, v.planeBBase(), 0, &planeBIdx, &planeBPri)
	v.renderPlaneLine(line, v.planeABase(), 1, &planeAIdx, &planeAPri)

	var spriteIdx [Width]byte
	var spritePri, spriteOpaque [Width]bool
	v.renderSpritesLine(line, &spriteIdx, &spritePri, &spriteOpaque)

	bg := v.backdropColor()
	for x := 0; x < Width; x++ {
		finalIdx := bg
		switch {
		case spriteOpaque[x] && spritePri[x]:
			finalIdx = spriteIdx[x]
		case planeAPri[x] && planeAIdx[x]&0x0F != 0:
			finalIdx = planeAIdx[x]
		case planeBPri[x] && planeBIdx[x]&0x0F != 0:
			finalIdx = planeBIdx[x]
		case spriteOpaque[x]:
			finalIdx = spriteIdx[x]
		case planeAIdx[x]&0x0F != 0:
			finalIdx = planeAIdx[x]
		case planeBIdx[x]&0x0F != 0:
			finalIdx = planeBIdx[x]
		}
		r, g, b := v.palette.rgb(finalIdx)
		v.frame.SetPixel(x, line, r, g, b, 0xFF)
	}
}
