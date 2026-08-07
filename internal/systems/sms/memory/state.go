package memory

// Snapshot captures the bus's own state: system RAM, the mapper's bank
// registers and optional cartridge RAM, and the joypad's held-button
// state (kept for consistency with the rest of a save, even though the
// frontend re-supplies it every frame during actual play).
type Snapshot struct {
	RAM [0x2000]byte

	MapperRAM           [0x2000]byte
	Slot0, Slot1, Slot2 byte
	RAMMapped           bool

	JoypadButtons uint32
}

func (b *Bus) Snapshot() Snapshot {
	return Snapshot{
		RAM:       b.ram,
		MapperRAM: b.mapper.ram,
		Slot0:     b.mapper.slot0, Slot1: b.mapper.slot1, Slot2: b.mapper.slot2,
		RAMMapped:     b.mapper.ramMapped,
		JoypadButtons: b.pad.state.Buttons,
	}
}

func (b *Bus) Restore(s Snapshot) {
	b.ram = s.RAM
	b.mapper.ram = s.MapperRAM
	b.mapper.slot0, b.mapper.slot1, b.mapper.slot2 = s.Slot0, s.Slot1, s.Slot2
	b.mapper.ramMapped = s.RAMMapped
	b.pad.state.Buttons = s.JoypadButtons
}
