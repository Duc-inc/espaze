package ppu

import "github.com/Duc-inc/espaze/internal/emulation/video"

// Screen dimensions - identical to DMG, CGB never changed resolution.
const (
	Width  = 160
	Height = 144
)

// Mode timing in T-cycles, one full scanline is 456 - identical to DMG
// in single-speed mode (the CPU's double-speed mode halves its own
// instruction timing, not the PPU's, which always runs at the same
// real-time rate regardless of CPU speed).
const (
	oamCycles      = 80
	transferCycles = 172
	hblankCycles   = 204
	lineCycles     = 456
)

const (
	modeHBlank   = 0
	modeVBlank   = 1
	modeOAM      = 2
	modeTransfer = 3
)

// Interrupt bits, matching memory.InterruptVBlank/InterruptSTAT so PPU
// doesn't need to import the memory package just for two constants.
const (
	interruptVBlank = 1 << 0
	interruptSTAT   = 1 << 1
)

// PPU renders the background, window and sprites into a 160x144 frame
// in full color - the CGB extension of the DMG PPU, with two VRAM banks
// (tile data in bank 0, tile *attributes* in bank 1 at the same
// addresses) and 8 background + 8 object palettes instead of the DMG's
// 4-shade BGP/OBP0/OBP1.
type PPU struct {
	vram     [2][0x2000]byte
	vramBank byte
	oam      [0xA0]byte

	bgPalettes, objPalettes cgbPalettes

	lcdc, stat byte
	scy, scx   byte
	ly, lyc    byte
	wy, wx     byte

	mode      byte
	modeClock int

	frame *video.FrameBuffer
}

// New returns a PPU with the LCD off and a blank frame, matching the
// state a cartridge's init code expects to find at boot.
func New() *PPU {
	return &PPU{frame: video.NewFrameBuffer(Width, Height)}
}

// Reset returns the PPU to its post-boot state, keeping VRAM/OAM as-is
// (LoadROM callers clear those separately if they need to).
func (p *PPU) Reset() {
	frame := p.frame
	*p = PPU{frame: frame}
}

// ReadVRAM/WriteVRAM implement CPU access to 0x8000-0x9FFF, through
// whichever bank VBK ($FF4F) currently selects.
func (p *PPU) ReadVRAM(addr uint16) byte     { return p.vram[p.vramBank][addr-0x8000] }
func (p *PPU) WriteVRAM(addr uint16, v byte) { p.vram[p.vramBank][addr-0x8000] = v }

// ReadOAM/WriteOAM implement CPU (and DMA) access to 0xFE00-0xFE9F.
func (p *PPU) ReadOAM(addr uint16) byte     { return p.oam[addr-0xFE00] }
func (p *PPU) WriteOAM(addr uint16, v byte) { p.oam[addr-0xFE00] = v }

// FrameBuffer returns the most recently completed frame.
func (p *PPU) FrameBuffer() *video.FrameBuffer { return p.frame }

// Step advances the PPU's mode state machine by tcycles and returns any
// interrupts it wants requested this step. Identical structure to the
// DMG PPU's Step - see internal/systems/gameboy/ppu/ppu.go - the only
// CGB-specific change is what renderScanline actually draws.
func (p *PPU) Step(tcycles int) byte {
	if p.lcdc&lcdcEnable == 0 {
		p.mode, p.ly, p.modeClock = modeHBlank, 0, 0
		return 0
	}

	var interrupts byte
	p.modeClock += tcycles

	switch p.mode {
	case modeOAM:
		if p.modeClock >= oamCycles {
			p.modeClock -= oamCycles
			p.mode = modeTransfer
		}
	case modeTransfer:
		if p.modeClock >= transferCycles {
			p.modeClock -= transferCycles
			p.mode = modeHBlank
			p.renderScanline()
			if p.stat&statHBlankEnable != 0 {
				interrupts |= interruptSTAT
			}
		}
	case modeHBlank:
		if p.modeClock >= hblankCycles {
			p.modeClock -= hblankCycles
			p.advanceLine(&interrupts)
			if p.ly == Height {
				p.mode = modeVBlank
				interrupts |= interruptVBlank
				if p.stat&statVBlankEnable != 0 {
					interrupts |= interruptSTAT
				}
			} else {
				p.mode = modeOAM
				if p.stat&statOAMEnable != 0 {
					interrupts |= interruptSTAT
				}
			}
		}
	case modeVBlank:
		if p.modeClock >= lineCycles {
			p.modeClock -= lineCycles
			p.advanceLine(&interrupts)
			if p.ly > 153 {
				p.ly = 0
				p.mode = modeOAM
				if p.stat&statOAMEnable != 0 {
					interrupts |= interruptSTAT
				}
			}
		}
	}

	p.stat = (p.stat &^ 0x03) | p.mode
	return interrupts
}

func (p *PPU) advanceLine(interrupts *byte) {
	p.ly++
	wasMatch := p.stat&statLYCFlag != 0
	isMatch := p.ly == p.lyc
	if isMatch {
		p.stat |= statLYCFlag
	} else {
		p.stat &^= statLYCFlag
	}
	if isMatch && !wasMatch && p.stat&statLYCEnable != 0 {
		*interrupts |= interruptSTAT
	}
}

func (p *PPU) renderScanline() {
	line := int(p.ly)
	var bg [Width]bgPixel

	p.renderBackgroundLine(line, &bg)
	if p.lcdc&lcdcWindowEnable != 0 {
		p.renderWindowLine(line, &bg)
	}
	p.renderSpritesLine(line, &bg)
}

// tileDataAddr resolves a tile index to its VRAM address per LCDC's
// addressing mode: unsigned from 0x8000, or signed from 0x9000.
func tileDataAddr(tileIndex byte, unsigned bool) uint16 {
	if unsigned {
		return 0x8000 + uint16(tileIndex)*16
	}
	return uint16(int32(0x9000) + int32(int8(tileIndex))*16)
}

// tilePixel reads one pixel's 2-bit color index out of an 8x8 tile,
// from the given VRAM bank (a CGB tile attribute can pull its pattern
// data from either bank, independent of which bank holds the tile map).
func (p *PPU) tilePixel(bank byte, tileAddr uint16, x, y int) byte {
	rowAddr := tileAddr + uint16(y)*2
	lo := p.vram[bank][rowAddr-0x8000]
	hi := p.vram[bank][rowAddr+1-0x8000]
	bit := 7 - x
	return ((hi>>bit)&1)<<1 | ((lo >> bit) & 1)
}

func (p *PPU) setPixel(x, y int, r, g, b byte) {
	p.frame.SetPixel(x, y, r, g, b, 0xFF)
}
