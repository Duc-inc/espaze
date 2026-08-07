package vdc

// Snapshot captures the VDC's VRAM, sprite attribute table, and every
// register. The frame buffer and the PaletteResolver are reinstated
// by the owning core, like every other video chip here.
type Snapshot struct {
	VRAM         [0x8000]uint16
	SAT          [64 * 4]uint16
	SelectedReg  byte
	Regs         [20]uint16
	WriteHiNext  bool
	VRAMLowLatch byte
	Line         int
	LineClock    int
}

// Snapshot captures the VDC's current state.
func (v *VDC) Snapshot() Snapshot {
	return Snapshot{
		VRAM: v.vram, SAT: v.sat, SelectedReg: v.selectedReg, Regs: v.regs,
		WriteHiNext: v.writeHiNext, VRAMLowLatch: v.vramLowLatch,
		Line: v.line, LineClock: v.lineClock,
	}
}

// Restore reinstates a previously captured Snapshot.
func (v *VDC) Restore(s Snapshot) {
	v.vram = s.VRAM
	v.sat = s.SAT
	v.selectedReg = s.SelectedReg
	v.regs = s.Regs
	v.writeHiNext = s.WriteHiNext
	v.vramLowLatch = s.VRAMLowLatch
	v.line = s.Line
	v.lineClock = s.LineClock
}
