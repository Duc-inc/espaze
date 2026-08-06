package memory

// mbc0 is a plain, unbanked cartridge: at most 32KB of ROM mapped straight
// through, with either no external RAM or a single fixed 8KB bank.
type mbc0 struct {
	rom []byte
	ram []byte
}

func newMBC0(cart *Cartridge) *mbc0 {
	return &mbc0{rom: cart.ROM, ram: make([]byte, max(cart.RAMSize, 0))}
}

func (m *mbc0) ReadROM(addr uint16) byte {
	if int(addr) >= len(m.rom) {
		return 0xFF
	}
	return m.rom[addr]
}

func (m *mbc0) WriteROM(uint16, byte) {} // no control registers

func (m *mbc0) ReadRAM(addr uint16) byte {
	idx := int(addr - 0xA000)
	if idx < 0 || idx >= len(m.ram) {
		return 0xFF
	}
	return m.ram[idx]
}

func (m *mbc0) WriteRAM(addr uint16, v byte) {
	idx := int(addr - 0xA000)
	if idx < 0 || idx >= len(m.ram) {
		return
	}
	m.ram[idx] = v
}

func (m *mbc0) Snapshot() MBCSnapshot {
	ram := make([]byte, len(m.ram))
	copy(ram, m.ram)
	return MBCSnapshot{RAM: ram}
}

func (m *mbc0) Restore(s MBCSnapshot) {
	copy(m.ram, s.RAM)
}
