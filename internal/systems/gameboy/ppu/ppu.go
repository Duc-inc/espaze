package ppu

import "github.com/Duc-inc/espaze/internal/emulation/video"

// Screen dimensions of the DMG LCD.
const (
	Width  = 160
	Height = 144
)

// Mode timing in T-cycles, one full scanline is 456.
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

// PPU renders the background, window and sprites into a 160x144 frame,
// one scanline at a time, driven by Step being called with however many
// T-cycles the CPU just spent.
type PPU struct {
	vram [0x2000]byte
	oam  [0xA0]byte

	lcdc, stat      byte
	scy, scx        byte
	ly, lyc         byte
	bgp, obp0, obp1 byte
	wy, wx          byte

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

// ReadVRAM/WriteVRAM implement CPU access to 0x8000-0x9FFF.
func (p *PPU) ReadVRAM(addr uint16) byte     { return p.vram[addr-0x8000] }
func (p *PPU) WriteVRAM(addr uint16, v byte) { p.vram[addr-0x8000] = v }

// ReadOAM/WriteOAM implement CPU (and DMA) access to 0xFE00-0xFE9F.
func (p *PPU) ReadOAM(addr uint16) byte     { return p.oam[addr-0xFE00] }
func (p *PPU) WriteOAM(addr uint16, v byte) { p.oam[addr-0xFE00] = v }

// FrameBuffer returns the most recently completed frame.
func (p *PPU) FrameBuffer() *video.FrameBuffer { return p.frame }

// Step advances the PPU's mode state machine by tcycles and returns any
// interrupts it wants requested this step.
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

// tileDataAddr resolves a tile index to its VRAM address per LCDC's
// addressing mode: unsigned from 0x8000, or signed from 0x9000.
func tileDataAddr(tileIndex byte, unsigned bool) uint16 {
	if unsigned {
		return 0x8000 + uint16(tileIndex)*16
	}
	return uint16(int32(0x9000) + int32(int8(tileIndex))*16)
}

// tilePixel reads one pixel's 2-bit color index out of an 8x8 tile.
func (p *PPU) tilePixel(tileAddr uint16, x, y int) byte {
	rowAddr := tileAddr + uint16(y)*2
	lo := p.vram[rowAddr-0x8000]
	hi := p.vram[rowAddr+1-0x8000]
	bit := 7 - x
	return ((hi>>bit)&1)<<1 | ((lo >> bit) & 1)
}

func applyPalette(colorIdx, palette byte) byte {
	return (palette >> (colorIdx * 2)) & 0x03
}

func (p *PPU) setPixel(x, y int, shade byte) {
	r, g, b := shadeToRGB(shade)
	p.frame.SetPixel(x, y, r, g, b, 0xFF)
}

func shadeToRGB(shade byte) (byte, byte, byte) {
	switch shade {
	case 0:
		return 0xE0, 0xE0, 0xE0
	case 1:
		return 0xA0, 0xA0, 0xA0
	case 2:
		return 0x60, 0x60, 0x60
	default:
		return 0x20, 0x20, 0x20
	}
}
