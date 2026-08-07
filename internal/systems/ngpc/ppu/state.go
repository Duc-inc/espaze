package ppu

// Snapshot captures the PPU's VRAM, sprite table, palette RAM, and
// every register. The frame buffer is reinstated by the owning core.
type Snapshot struct {
	VRAM                [0x4000]byte
	Sprites             [64 * 4]byte
	Palette             [32]uint16
	ScrollX, ScrollY    byte
	BGEnable, OBJEnable bool
	Line, LineClock     int
}

// Snapshot captures the PPU's current state.
func (p *PPU) Snapshot() Snapshot {
	return Snapshot{
		VRAM: p.vram, Sprites: p.sprites, Palette: p.palette,
		ScrollX: p.scrollX, ScrollY: p.scrollY,
		BGEnable: p.bgEnable, OBJEnable: p.objEnable,
		Line: p.line, LineClock: p.lineClock,
	}
}

// Restore reinstates a previously captured Snapshot.
func (p *PPU) Restore(s Snapshot) {
	p.vram, p.sprites, p.palette = s.VRAM, s.Sprites, s.Palette
	p.scrollX, p.scrollY = s.ScrollX, s.ScrollY
	p.bgEnable, p.objEnable = s.BGEnable, s.OBJEnable
	p.line, p.lineClock = s.Line, s.LineClock
}
