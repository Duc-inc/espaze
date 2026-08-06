package ppu

// LCDC bits (0xFF40). Identical layout to DMG, except bit0's meaning
// changes in CGB mode: instead of "background on/off", it becomes
// "master priority" - when clear, sprites always draw on top of the
// background/window regardless of any per-tile or per-sprite priority
// bit (see compositePixel in sprites.go).
const (
	lcdcEnable         = 1 << 7
	lcdcWindowMap      = 1 << 6
	lcdcWindowEnable   = 1 << 5
	lcdcTileData       = 1 << 4
	lcdcBGMap          = 1 << 3
	lcdcOBJSize        = 1 << 2
	lcdcOBJEnable      = 1 << 1
	lcdcMasterPriority = 1 << 0
)

// STAT interrupt-source bits (0xFF41).
const (
	statLYCEnable    = 1 << 6
	statOAMEnable    = 1 << 5
	statVBlankEnable = 1 << 4
	statHBlankEnable = 1 << 3
	statLYCFlag      = 1 << 2
)

// ReadRegister implements the CPU reading 0xFF40-0xFF4B plus the CGB
// additions (0xFF4F VBK, 0xFF68-0xFF6B palette ports). 0xFF46 (OAM DMA)
// is intercepted by the MMU before it ever reaches here.
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
	case 0xFF4A:
		return p.wy
	case 0xFF4B:
		return p.wx
	case 0xFF4F:
		return p.vramBank | 0xFE
	case 0xFF68:
		return p.bgPalettes.readIndexPort()
	case 0xFF69:
		return p.bgPalettes.readData()
	case 0xFF6A:
		return p.objPalettes.readIndexPort()
	case 0xFF6B:
		return p.objPalettes.readData()
	default:
		return 0xFF
	}
}

// WriteRegister implements the CPU writing to the same range.
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
	case 0xFF4A:
		p.wy = v
	case 0xFF4B:
		p.wx = v
	case 0xFF4F:
		p.vramBank = v & 0x01
	case 0xFF68:
		p.bgPalettes.writeIndex(v)
	case 0xFF69:
		p.bgPalettes.writeData(v)
	case 0xFF6A:
		p.objPalettes.writeIndex(v)
	case 0xFF6B:
		p.objPalettes.writeData(v)
	}
	// 0xFF44 (LY) is read-only.
}
