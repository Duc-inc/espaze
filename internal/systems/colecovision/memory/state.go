package memory

// Snapshot captures the bus's own state: RAM and the controller's
// last-seen buttons. The cartridge ROM is never included, and the
// VDP/PSG are snapshotted separately by their owners.
type Snapshot struct {
	RAM     [0x0400]byte
	PadBtns byte
}

// Snapshot captures the bus's current state.
func (b *Bus) Snapshot() Snapshot { return Snapshot{RAM: b.ram, PadBtns: b.pad.buttons} }

// Restore reinstates a previously captured Snapshot.
func (b *Bus) Restore(s Snapshot) {
	b.ram = s.RAM
	b.pad.buttons = s.PadBtns
}
