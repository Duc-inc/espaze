package memory

// Snapshot captures the bus's own state: work RAM, the OAM address
// latch, and the controller's last-seen buttons. The cartridge ROM is
// never included, and the PPU/audio ports are snapshotted separately
// by their owners.
type Snapshot struct {
	WRAM    [0x20000]byte
	OAMAddr uint16
	PadBtns uint16
}

// Snapshot captures the bus's current state.
func (b *Bus) Snapshot() Snapshot {
	return Snapshot{WRAM: b.wram, OAMAddr: b.oamAddr, PadBtns: b.pad1.buttons}
}

// Restore reinstates a previously captured Snapshot.
func (b *Bus) Restore(s Snapshot) {
	b.wram = s.WRAM
	b.oamAddr = s.OAMAddr
	b.pad1.buttons = s.PadBtns
}
