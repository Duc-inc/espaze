package ppu

// LCDC bits (0xFF40).
const (
	lcdcEnable       = 1 << 7
	lcdcWindowMap    = 1 << 6
	lcdcWindowEnable = 1 << 5
	lcdcTileData     = 1 << 4
	lcdcBGMap        = 1 << 3
	lcdcOBJSize      = 1 << 2
	lcdcOBJEnable    = 1 << 1
	lcdcBGEnable     = 1 << 0
)

// STAT interrupt-source bits (0xFF41).
const (
	statLYCEnable   = 1 << 6
	statOAMEnable   = 1 << 5
	statVBlankEnable = 1 << 4
	statHBlankEnable = 1 << 3
	statLYCFlag     = 1 << 2
)

// ReadRegister implements the CPU reading 0xFF40-0xFF4B (0xFF46/DMA is
// intercepted by the MMU before it ever reaches here).
func (p *PPU) ReadRegister(addr uint16) byte {
	switch addr {
	case 0xFF40:
		return p.lcdc
	case 0xFF41:
		return p.stat | 0x80
	case 0xFF42:
		return p.scy
	case 0xFF43:
		return p.scx
	case 0xFF44:
		return p.ly
	case 0xFF45:
		return p.lyc
	case 0xFF47:
		return p.bgp
	case 0xFF48:
		return p.obp0
	case 0xFF49:
		return p.obp1
	case 0xFF4A:
		return p.wy
	case 0xFF4B:
		return p.wx
	default:
		return 0xFF
	}
}

// WriteRegister implements the CPU writing 0xFF40-0xFF4B.
func (p *PPU) WriteRegister(addr uint16, v byte) {
	switch addr {
	case 0xFF40:
		p.lcdc = v
	case 0xFF41:
		p.stat = (p.stat & 0x07) | (v &^ 0x07)
	case 0xFF42:
		p.scy = v
	case 0xFF43:
		p.scx = v
	case 0xFF45:
		p.lyc = v
	case 0xFF47:
		p.bgp = v
	case 0xFF48:
		p.obp0 = v
	case 0xFF49:
		p.obp1 = v
	case 0xFF4A:
		p.wy = v
	case 0xFF4B:
		p.wx = v
	}
	// 0xFF44 (LY) is read-only.
}
