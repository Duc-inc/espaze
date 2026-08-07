package memory

// Snapshot captures the 68000-side bus's own state: work RAM, the
// controller's latched TH line, and the Z80 reset line. The cartridge
// ROM is never included, and the VDP/Z80/audio chips are snapshotted
// separately by their owners, matching every other core in this
// project.
type Snapshot struct {
	WRAM     [0x10000]byte
	Pad1TH   bool
	Pad1Btns byte
	Z80Reset bool
}

// Snapshot captures the bus's current state.
func (b *Bus) Snapshot() Snapshot {
	return Snapshot{
		WRAM:     b.wram,
		Pad1TH:   b.pad1.th,
		Pad1Btns: b.pad1.buttons,
		Z80Reset: b.z80Reset,
	}
}

// Restore reinstates a previously captured Snapshot.
func (b *Bus) Restore(s Snapshot) {
	b.wram = s.WRAM
	b.pad1.th = s.Pad1TH
	b.pad1.buttons = s.Pad1Btns
	b.z80Reset = s.Z80Reset
}
