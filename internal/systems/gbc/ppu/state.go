package ppu

// Snapshot captures every piece of PPU state.
type Snapshot struct {
	VRAM     [2][0x2000]byte
	VRAMBank byte
	OAM      [0xA0]byte

	BGPalette, OBJPalette   [64]byte
	BGPalIndex, BGPalAuto   byte
	ObjPalIndex, ObjPalAuto byte

	LCDC, Stat byte
	SCY, SCX   byte
	LY, LYC    byte
	WY, WX     byte

	Mode      byte
	ModeClock int
}

func (p *PPU) Snapshot() Snapshot {
	bgAuto, objAuto := byte(0), byte(0)
	if p.bgPalettes.autoIncr {
		bgAuto = 1
	}
	if p.objPalettes.autoIncr {
		objAuto = 1
	}
	return Snapshot{
		VRAM: p.vram, VRAMBank: p.vramBank, OAM: p.oam,
		BGPalette: p.bgPalettes.data, OBJPalette: p.objPalettes.data,
		BGPalIndex: p.bgPalettes.index, BGPalAuto: bgAuto,
		ObjPalIndex: p.objPalettes.index, ObjPalAuto: objAuto,
		LCDC: p.lcdc, Stat: p.stat,
		SCY: p.scy, SCX: p.scx, LY: p.ly, LYC: p.lyc, WY: p.wy, WX: p.wx,
		Mode: p.mode, ModeClock: p.modeClock,
	}
}

func (p *PPU) Restore(s Snapshot) {
	p.vram, p.vramBank, p.oam = s.VRAM, s.VRAMBank, s.OAM
	p.bgPalettes.data, p.objPalettes.data = s.BGPalette, s.OBJPalette
	p.bgPalettes.index, p.bgPalettes.autoIncr = s.BGPalIndex, s.BGPalAuto != 0
	p.objPalettes.index, p.objPalettes.autoIncr = s.ObjPalIndex, s.ObjPalAuto != 0
	p.lcdc, p.stat = s.LCDC, s.Stat
	p.scy, p.scx, p.ly, p.lyc, p.wy, p.wx = s.SCY, s.SCX, s.LY, s.LYC, s.WY, s.WX
	p.mode, p.modeClock = s.Mode, s.ModeClock
}
