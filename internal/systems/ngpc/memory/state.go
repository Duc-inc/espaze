package memory

// Snapshot captures the bus's own state: work RAM, the controller's
// last-seen buttons, and the Z80 reset line. The cartridge ROM is
// never included, and the PPU/audio bus are snapshotted separately by
// their owners.
type Snapshot struct {
	WRAM     [0x4000]byte
	PadBtns  byte
	Z80Reset bool
}

// Snapshot captures the bus's current state.
func (b *Bus) Snapshot() Snapshot {
	return Snapshot{WRAM: b.wram, PadBtns: b.pad.buttons, Z80Reset: b.z80Reset}
}

// Restore reinstates a previously captured Snapshot.
func (b *Bus) Restore(s Snapshot) {
	b.wram = s.WRAM
	b.pad.buttons = s.PadBtns
	b.z80Reset = s.Z80Reset
}
