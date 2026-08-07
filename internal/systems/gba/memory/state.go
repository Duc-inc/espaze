package memory

// dmaSnapshot/timerSnapshot flatten dmaChannel/timer's unexported
// fields into exported ones - gob silently refuses to encode a struct
// whose fields are all unexported, so every snapshot in this project
// does this rather than embedding the internal type directly.
type dmaSnapshot struct {
	Src, Dst uint32
	Count    uint16
	Control  uint16
}

func snapshotDMA(d *dmaChannel) dmaSnapshot {
	return dmaSnapshot{Src: d.src, Dst: d.dst, Count: d.count, Control: d.control}
}

func restoreDMA(d *dmaChannel, s dmaSnapshot) {
	d.src, d.dst, d.count, d.control = s.Src, s.Dst, s.Count, s.Control
}

type timerSnapshot struct {
	Counter, Reload uint16
	Prescaler       int
	Cascade         bool
	IRQEnable       bool
	Running         bool
	Sub             int
}

func snapshotTimer(t *timer) timerSnapshot {
	return timerSnapshot{
		Counter: t.counter, Reload: t.reload, Prescaler: t.prescaler,
		Cascade: t.cascade, IRQEnable: t.irqEnable, Running: t.running, Sub: t.sub,
	}
}

func restoreTimer(t *timer, s timerSnapshot) {
	t.counter, t.reload, t.prescaler = s.Counter, s.Reload, s.Prescaler
	t.cascade, t.irqEnable, t.running, t.sub = s.Cascade, s.IRQEnable, s.Running, s.Sub
}

// Snapshot captures the bus's own state: both WRAM regions, SRAM,
// DMA channels, timers, the keypad, and interrupt controller. The
// cartridge ROM is never included, and the PPU/APU are snapshotted
// separately by their owners, matching every other core in this project.
type Snapshot struct {
	EWRAM         [0x40000]byte
	IWRAM         [0x8000]byte
	SRAM          [0x8000]byte
	DMA           [4]dmaSnapshot
	Timers        [4]timerSnapshot
	KeypadButtons uint16
	IE, IF        uint16
	IME           bool
}

// Snapshot captures the bus's current state.
func (b *Bus) Snapshot() Snapshot {
	var dma [4]dmaSnapshot
	var tms [4]timerSnapshot
	for i := range b.dma {
		dma[i] = snapshotDMA(&b.dma[i])
	}
	for i := range b.tm.t {
		tms[i] = snapshotTimer(&b.tm.t[i])
	}
	return Snapshot{
		EWRAM: b.ewram, IWRAM: b.iwram, SRAM: b.sram.data,
		DMA: dma, Timers: tms, KeypadButtons: b.kp.buttons,
		IE: b.irq.ie, IF: b.irq.iflags, IME: b.irq.ime,
	}
}

// Restore reinstates a previously captured Snapshot.
func (b *Bus) Restore(s Snapshot) {
	b.ewram, b.iwram, b.sram.data = s.EWRAM, s.IWRAM, s.SRAM
	for i := range b.dma {
		restoreDMA(&b.dma[i], s.DMA[i])
	}
	for i := range b.tm.t {
		restoreTimer(&b.tm.t[i], s.Timers[i])
	}
	b.kp.buttons = s.KeypadButtons
	b.irq.ie, b.irq.iflags, b.irq.ime = s.IE, s.IF, s.IME
}
