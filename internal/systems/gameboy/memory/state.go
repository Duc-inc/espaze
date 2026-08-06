package memory

// Snapshot captures everything the MMU itself owns directly (not the
// cartridge or the APU, which callers snapshot separately via
// MBC.Snapshot and apu.APU.Snapshot).
type Snapshot struct {
	WRAM   [0x2000]byte
	HRAM   [0x7F]byte
	IFReg  byte
	IEReg  byte
	Serial [2]byte
}

func (m *MMU) Snapshot() Snapshot {
	return Snapshot{
		WRAM: m.wram, HRAM: m.hram,
		IFReg: m.ifReg, IEReg: m.ieReg,
		Serial: m.serial,
	}
}

func (m *MMU) Restore(s Snapshot) {
	m.wram, m.hram = s.WRAM, s.HRAM
	m.ifReg, m.ieReg = s.IFReg, s.IEReg
	m.serial = s.Serial
}
