package vdp

// Snapshot captures the VDP's full state: VRAM/CRAM/VSRAM, every
// register, and the control-port protocol's in-flight state. The frame
// buffer and the DMA MemoryReader are never included - both are
// reinstated by the owning core, exactly like every other video chip
// in this project.
type Snapshot struct {
	VRAM         [0x10000]byte
	CRAM         [64]uint16
	VSRAM        [40]uint16
	Regs         [24]byte
	Code         byte
	Addr         uint16
	CtrlLow      uint16
	CtrlPending  bool
	DMAFillArmed bool
	Status       byte
	LineClock    int
	Line         int
}

// Snapshot captures the VDP's current state.
func (v *VDP) Snapshot() Snapshot {
	return Snapshot{
		VRAM:         v.vram,
		CRAM:         v.palette.data,
		VSRAM:        v.vsram,
		Regs:         v.regs,
		Code:         v.code,
		Addr:         v.addr,
		CtrlLow:      v.ctrlLow,
		CtrlPending:  v.ctrlPending,
		DMAFillArmed: v.dmaFillArmed,
		Status:       v.status,
		LineClock:    v.lineClock,
		Line:         v.line,
	}
}

// Restore reinstates a previously captured Snapshot.
func (v *VDP) Restore(s Snapshot) {
	v.vram = s.VRAM
	v.palette.data = s.CRAM
	v.vsram = s.VSRAM
	v.regs = s.Regs
	v.code = s.Code
	v.addr = s.Addr
	v.ctrlLow = s.CtrlLow
	v.ctrlPending = s.CtrlPending
	v.dmaFillArmed = s.DMAFillArmed
	v.status = s.Status
	v.lineClock = s.LineClock
	v.line = s.Line
}
