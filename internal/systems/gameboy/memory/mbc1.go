package memory

// mbc1 implements the most common Game Boy memory bank controller: up to
// 125 usable 16KB ROM banks and up to four 8KB RAM banks, selected via
// writes to ROM address space (the cartridge intercepts them, nothing is
// actually stored there).
type mbc1 struct {
	rom []byte
	ram []byte

	ramEnabled  bool
	romBankLow5 uint8
	upperBits   uint8 // RAM bank, or high ROM bank bits depending on mode
	bankingMode uint8 // 0 = ROM banking, 1 = RAM banking
}

func newMBC1(cart *Cartridge) *mbc1 {
	return &mbc1{
		rom:         cart.ROM,
		ram:         make([]byte, max(cart.RAMSize, 0)),
		romBankLow5: 1,
	}
}

func (m *mbc1) WriteROM(addr uint16, v byte) {
	switch {
	case addr <= 0x1FFF:
		m.ramEnabled = v&0x0F == 0x0A
	case addr <= 0x3FFF:
		bank := v & 0x1F
		if bank == 0 {
			bank = 1
		}
		m.romBankLow5 = bank
	case addr <= 0x5FFF:
		m.upperBits = v & 0x03
	default: // 0x6000-0x7FFF
		m.bankingMode = v & 0x01
	}
}

func (m *mbc1) ReadROM(addr uint16) byte {
	var offset int
	if addr < 0x4000 {
		offset = int(addr) // bank 0, fixed
	} else {
		bank := m.romBankLow5
		if m.bankingMode == 0 {
			bank |= m.upperBits << 5
		}
		offset = int(bank)*0x4000 + int(addr-0x4000)
	}
	if offset >= len(m.rom) {
		return 0xFF
	}
	return m.rom[offset]
}

func (m *mbc1) ramOffset(addr uint16) int {
	bank := 0
	if m.bankingMode == 1 {
		bank = int(m.upperBits)
	}
	return bank*0x2000 + int(addr-0xA000)
}

func (m *mbc1) ReadRAM(addr uint16) byte {
	if !m.ramEnabled {
		return 0xFF
	}
	idx := m.ramOffset(addr)
	if idx < 0 || idx >= len(m.ram) {
		return 0xFF
	}
	return m.ram[idx]
}

func (m *mbc1) WriteRAM(addr uint16, v byte) {
	if !m.ramEnabled {
		return
	}
	idx := m.ramOffset(addr)
	if idx < 0 || idx >= len(m.ram) {
		return
	}
	m.ram[idx] = v
}

func (m *mbc1) Snapshot() MBCSnapshot {
	ram := make([]byte, len(m.ram))
	copy(ram, m.ram)
	return MBCSnapshot{
		RAM:         ram,
		ROMBank:     m.romBankLow5,
		RAMBank:     m.upperBits,
		RAMEnabled:  m.ramEnabled,
		BankingMode: m.bankingMode,
	}
}

func (m *mbc1) Restore(s MBCSnapshot) {
	copy(m.ram, s.RAM)
	m.romBankLow5 = s.ROMBank
	m.upperBits = s.RAMBank
	m.ramEnabled = s.RAMEnabled
	m.bankingMode = s.BankingMode
}
