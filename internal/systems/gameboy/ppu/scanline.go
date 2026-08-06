package ppu

// renderScanline paints one full line (background, window, sprites) into
// the frame buffer, called once per scanline as Transfer mode ends.
func (p *PPU) renderScanline() {
	line := int(p.ly)
	if line < 0 || line >= Height {
		return
	}

	var bgIdx [Width]byte

	if p.lcdc&lcdcBGEnable != 0 {
		p.renderBackgroundLine(line, &bgIdx)
		if p.lcdc&lcdcWindowEnable != 0 {
			p.renderWindowLine(line, &bgIdx)
		}
	} else {
		for x := 0; x < Width; x++ {
			p.setPixel(x, line, 0)
		}
	}

	if p.lcdc&lcdcOBJEnable != 0 {
		p.renderSpritesLine(line, &bgIdx)
	}
}

// Snapshot/Restore persist every register plus VRAM/OAM for save states.
type Snapshot struct {
	VRAM [0x2000]byte
	OAM  [0xA0]byte

	LCDC, Stat     byte
	SCY, SCX       byte
	LY, LYC        byte
	BGP, OBP0, OBP1 byte
	WY, WX         byte

	Mode      byte
	ModeClock int

	// FramePixels is the last rendered picture, saved too so the screen
	// shows something correct immediately on load instead of staying
	// blank until the next full frame finishes rendering.
	FramePixels []byte
}

func (p *PPU) Snapshot() Snapshot {
	return Snapshot{
		VRAM: p.vram, OAM: p.oam,
		LCDC: p.lcdc, Stat: p.stat,
		SCY: p.scy, SCX: p.scx,
		LY: p.ly, LYC: p.lyc,
		BGP: p.bgp, OBP0: p.obp0, OBP1: p.obp1,
		WY: p.wy, WX: p.wx,
		Mode: p.mode, ModeClock: p.modeClock,
		FramePixels: append([]byte(nil), p.frame.Pixels...),
	}
}

func (p *PPU) Restore(s Snapshot) {
	p.vram, p.oam = s.VRAM, s.OAM
	p.lcdc, p.stat = s.LCDC, s.Stat
	p.scy, p.scx = s.SCY, s.SCX
	p.ly, p.lyc = s.LY, s.LYC
	p.bgp, p.obp0, p.obp1 = s.BGP, s.OBP0, s.OBP1
	p.wy, p.wx = s.WY, s.WX
	p.mode, p.modeClock = s.Mode, s.ModeClock
	if len(s.FramePixels) == len(p.frame.Pixels) {
		copy(p.frame.Pixels, s.FramePixels)
	}
}
