package ppu

import "github.com/Duc-inc/espaze/internal/emulation/video"

// Screen dimensions of the NES's visible picture.
const (
	Width  = 256
	Height = 240
)

// Scanline/dot timing (NTSC): 341 dots per scanline, 262 scanlines per
// frame, 3 PPU dots per CPU cycle.
const (
	dotsPerScanline     = 341
	visibleScanlines    = 240
	vblankStartScanline = 241
	preRenderScanline   = 261
	lastScanline        = 261
)

// PPU renders the background and sprites into a 256x240 frame, one
// scanline at a time (a batch equivalent of the real hardware's
// per-dot tile-fetch pipeline, not a cycle-exact re-implementation of
// it - see renderScanline). Pattern table data lives on the cartridge
// behind CartBus; everything else here is PPU-internal.
type PPU struct {
	nametables [4][1024]byte
	palette    paletteRAM
	oamMem     oam
	scroll     scrollRegs

	ctrl, mask, status byte
	oamAddr            byte
	dataBuffer         byte

	cart CartBus

	dot, scanline int
	oddFrame      bool

	frame *video.FrameBuffer
}

// New builds a PPU wired to cart (pattern tables and mirroring).
func New(cart CartBus) *PPU {
	return &PPU{cart: cart, frame: video.NewFrameBuffer(Width, Height)}
}

// Reset clears all PPU state except the cartridge link and the frame
// buffer (keeping the same *video.FrameBuffer instance since callers
// may already be holding a reference to it).
func (p *PPU) Reset() {
	cart, frame := p.cart, p.frame
	*p = PPU{cart: cart, frame: frame}
}

// FrameBuffer returns the most recently rendered picture.
func (p *PPU) FrameBuffer() *video.FrameBuffer { return p.frame }

func (p *PPU) mirroring() MirrorMode { return p.cart.Mirroring() }

func (p *PPU) renderingEnabled() bool { return p.mask&(maskShowBG|maskShowSprites) != 0 }

// WriteOAMByte lets $4014 OAM DMA (implemented at the bus level, since
// it also has to stall the CPU for 513-514 cycles) drop bytes into OAM
// starting at the current OAMADDR, exactly like repeated OAMDATA writes.
func (p *PPU) WriteOAMByte(v byte) {
	p.oamMem.writeByte(p.oamAddr, v)
	p.oamAddr++
}

// Step advances the PPU by ppuDots dots (3 per CPU cycle on NTSC) and
// reports whether it asserted NMI at any point during that span.
func (p *PPU) Step(ppuDots int) bool {
	nmi := false
	for i := 0; i < ppuDots; i++ {
		if p.tick() {
			nmi = true
		}
	}
	return nmi
}

func (p *PPU) tick() bool {
	nmi := false

	if p.scanline < visibleScanlines && p.dot == 0 {
		p.renderScanline(p.scanline)
	}

	onRenderedLine := p.scanline < visibleScanlines || p.scanline == preRenderScanline
	if onRenderedLine {
		switch {
		case p.dot == 256:
			p.scroll.incrementY()
		case p.dot == 257:
			p.scroll.copyHorizontal()
		case p.scanline == preRenderScanline && p.dot >= 280 && p.dot <= 304:
			p.scroll.copyVertical()
		}
	}

	if p.scanline == preRenderScanline && p.dot == 1 {
		p.status &^= statusVBlank | statusSprite0Hit | statusSpriteOverflow
	}
	if p.scanline == vblankStartScanline && p.dot == 1 {
		p.status |= statusVBlank
		if p.ctrl&ctrlNMIEnable != 0 {
			nmi = true
		}
	}

	p.dot++
	if p.dot > 340 {
		p.dot = 0
		p.scanline++
		if p.scanline > lastScanline {
			p.scanline = 0
			p.oddFrame = !p.oddFrame
			if p.oddFrame && p.renderingEnabled() {
				p.dot = 1 // the well-known odd-frame dot skip
			}
		}
	}
	return nmi
}

// renderScanline resolves every pixel of screenY in one pass: the
// background line, the sprites that intersect it, then composites them
// with sprite priority and sprite-0 hit detection.
func (p *PPU) renderScanline(screenY int) {
	bg := p.renderBackgroundLine()

	height := 8
	if p.ctrl&ctrlSpriteSize8x16 != 0 {
		height = 16
	}
	sprites, overflow := evaluateSprites(&p.oamMem, screenY, height)
	if overflow {
		p.status |= statusSpriteOverflow
	}
	sprite := p.renderSpriteLine(sprites)

	showBG := p.mask&maskShowBG != 0
	showSprites := p.mask&maskShowSprites != 0

	for x := 0; x < 256; x++ {
		master := bg.master[x]
		if sprite[x].opaque {
			if sprite[x].isSprite0 && bg.rawColor[x] != 0 && showBG && showSprites && x != 255 {
				p.status |= statusSprite0Hit
			}
			if !sprite[x].behindBG || bg.rawColor[x] == 0 {
				master = sprite[x].master
			}
		}
		rgb := masterPalette[master&0x3F]
		p.frame.SetPixel(x, screenY, rgb[0], rgb[1], rgb[2], 0xFF)
	}
}
