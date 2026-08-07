package ppu

// Snapshot captures the PPU's VRAM, OAM, palette RAM, and every
// register. The frame buffer is reinstated by the owning core, like
// every other video chip in this project.
type Snapshot struct {
	VRAM      [0x18000]byte
	OAM       [0x400]byte
	Palette   [512]uint16
	DISPCNT   uint16
	DISPSTAT  uint16
	BGCNT     [4]uint16
	BGHOFS    [4]uint16
	BGVOFS    [4]uint16
	Dot, Line int
}

// Snapshot captures the PPU's current state.
func (p *PPU) Snapshot() Snapshot {
	return Snapshot{
		VRAM: p.vram, OAM: p.oam, Palette: p.palette,
		DISPCNT: p.dispcnt, DISPSTAT: p.dispstat,
		BGCNT: p.bgcnt, BGHOFS: p.bghofs, BGVOFS: p.bgvofs,
		Dot: p.dot, Line: p.line,
	}
}

// Restore reinstates a previously captured Snapshot.
func (p *PPU) Restore(s Snapshot) {
	p.vram, p.oam, p.palette = s.VRAM, s.OAM, s.Palette
	p.dispcnt, p.dispstat = s.DISPCNT, s.DISPSTAT
	p.bgcnt, p.bghofs, p.bgvofs = s.BGCNT, s.BGHOFS, s.BGVOFS
	p.dot, p.line = s.Dot, s.Line
}
