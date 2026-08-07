package tms9918

// Snapshot captures the chip's VRAM and every register. The frame
// buffer is reinstated by the owning core.
type Snapshot struct {
	VRAM       [0x4000]byte
	Addr       uint16
	AddrLatch  byte
	Latched    bool
	WriteMode  bool
	ReadBuffer byte
	Regs       [8]byte
	Status     byte
	Line       int
	LineClock  int
}

// Snapshot captures the chip's current state.
func (t *TMS9918) Snapshot() Snapshot {
	return Snapshot{
		VRAM: t.vram, Addr: t.addr, AddrLatch: t.addrLatch,
		Latched: t.latched, WriteMode: t.writeMode, ReadBuffer: t.readBuffer,
		Regs: t.regs, Status: t.status, Line: t.line, LineClock: t.lineClock,
	}
}

// Restore reinstates a previously captured Snapshot.
func (t *TMS9918) Restore(s Snapshot) {
	t.vram, t.addr, t.addrLatch = s.VRAM, s.Addr, s.AddrLatch
	t.latched, t.writeMode, t.readBuffer = s.Latched, s.WriteMode, s.ReadBuffer
	t.regs, t.status, t.line, t.lineClock = s.Regs, s.Status, s.Line, s.LineClock
}
