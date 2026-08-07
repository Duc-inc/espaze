package memory

// Snapshot captures the bus's own state: work RAM and the pad's
// latched SEL line. The cartridge ROM is never included, and the
// VDC/VCE/PSG/timer are snapshotted separately by their owners.
type Snapshot struct {
	RAM     [0x2000]byte
	PadSel  bool
	PadBtns byte
}

// Snapshot captures the bus's current state.
func (b *Bus) Snapshot() Snapshot {
	return Snapshot{RAM: b.ram, PadSel: b.pad.sel, PadBtns: b.pad.buttons}
}

// Restore reinstates a previously captured Snapshot.
func (b *Bus) Restore(s Snapshot) {
	b.ram = s.RAM
	b.pad.sel = s.PadSel
	b.pad.buttons = s.PadBtns
}
