package memory

// Snapshot captures everything the MMU itself owns directly (not the
// cartridge, PPU or APU, which callers snapshot separately).
type Snapshot struct {
	WRAMBanks [8][0x1000]byte
	WRAMBank  byte
	HRAM      [0x7F]byte

	HDMASrcHi, HDMASrcLo byte
	HDMADstHi, HDMADstLo byte

	DoubleSpeed bool

	IFReg  byte
	IEReg  byte
	Serial [2]byte
}

func (m *MMU) Snapshot() Snapshot {
	return Snapshot{
		WRAMBanks: m.wram.banks, WRAMBank: m.wram.bank, HRAM: m.hram,
		HDMASrcHi: m.hdma.srcHi, HDMASrcLo: m.hdma.srcLo,
		HDMADstHi: m.hdma.dstHi, HDMADstLo: m.hdma.dstLo,
		DoubleSpeed: m.speed.doubleSpeed,
		IFReg:       m.ifReg, IEReg: m.ieReg, Serial: m.serial,
	}
}

func (m *MMU) Restore(s Snapshot) {
	m.wram.banks, m.wram.bank, m.hram = s.WRAMBanks, s.WRAMBank, s.HRAM
	m.hdma.srcHi, m.hdma.srcLo = s.HDMASrcHi, s.HDMASrcLo
	m.hdma.dstHi, m.hdma.dstLo = s.HDMADstHi, s.HDMADstLo
	m.speed.doubleSpeed = s.DoubleSpeed
	m.ifReg, m.ieReg, m.serial = s.IFReg, s.IEReg, s.Serial
}
