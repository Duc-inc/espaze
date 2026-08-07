package vdp

// Snapshot captures every piece of VDP state.
type Snapshot struct {
	VRAM [0x4000]byte
	CRAM [32]byte
	Regs [11]byte

	Addr        uint16
	CtrlLow     byte
	CtrlLatched bool
	Mode        accessMode
	ReadBuffer  byte

	Status byte

	LineCounter    byte
	LineIRQPending bool

	LineClock int
	Line      int
}

func (v *VDP) Snapshot() Snapshot {
	return Snapshot{
		VRAM: v.vram, CRAM: v.cram.data, Regs: v.regs,
		Addr: v.addr, CtrlLow: v.ctrlLow, CtrlLatched: v.ctrlLatched,
		Mode: v.mode, ReadBuffer: v.readBuffer,
		Status:         v.status,
		LineCounter:    v.lineCounter,
		LineIRQPending: v.lineIRQPending,
		LineClock:      v.lineClock, Line: v.line,
	}
}

func (v *VDP) Restore(s Snapshot) {
	v.vram, v.cram.data, v.regs = s.VRAM, s.CRAM, s.Regs
	v.addr, v.ctrlLow, v.ctrlLatched = s.Addr, s.CtrlLow, s.CtrlLatched
	v.mode, v.readBuffer = s.Mode, s.ReadBuffer
	v.status = s.Status
	v.lineCounter = s.LineCounter
	v.lineIRQPending = s.LineIRQPending
	v.lineClock, v.line = s.LineClock, s.Line
}
