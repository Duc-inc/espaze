package ppu

// Snapshot captures the PPU's VRAM, OAM, CGRAM, and every register.
// The frame buffer is reinstated by the owning core.
type Snapshot struct {
	VRAM        [0x10000]uint16
	OAM         [0x220]byte
	CGRAM       [256]uint16
	BGCnt       [4]byte
	BGHOfs      [4]uint16
	BGVOfs      [4]uint16
	BGMain      byte
	OBJMain     bool
	VRAMAddr    uint16
	VRAMLowNext bool
	CGRAMAddr   byte
	CGRAMHigh   bool
	Line        int
	LineClock   int
}

// Snapshot captures the PPU's current state.
func (p *PPU) Snapshot() Snapshot {
	return Snapshot{
		VRAM: p.vram, OAM: p.oam, CGRAM: p.cgram,
		BGCnt: p.bgcnt, BGHOfs: p.bgHOfs, BGVOfs: p.bgVOfs,
		BGMain: p.bgMain, OBJMain: p.objMain,
		VRAMAddr: p.vramAddr, VRAMLowNext: p.vramLowNext,
		CGRAMAddr: p.cgramAddr, CGRAMHigh: p.cgramHigh,
		Line: p.line, LineClock: p.lineClock,
	}
}

// Restore reinstates a previously captured Snapshot.
func (p *PPU) Restore(s Snapshot) {
	p.vram, p.oam, p.cgram = s.VRAM, s.OAM, s.CGRAM
	p.bgcnt, p.bgHOfs, p.bgVOfs = s.BGCnt, s.BGHOfs, s.BGVOfs
	p.bgMain, p.objMain = s.BGMain, s.OBJMain
	p.vramAddr, p.vramLowNext = s.VRAMAddr, s.VRAMLowNext
	p.cgramAddr, p.cgramHigh = s.CGRAMAddr, s.CGRAMHigh
	p.line, p.lineClock = s.Line, s.LineClock
}
