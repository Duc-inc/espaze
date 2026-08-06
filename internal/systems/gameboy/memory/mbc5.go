package memory

// mbc5 supports up to 512 16KB ROM banks (a 9-bit bank number, unlike
// MBC1's 5) and up to 16 8KB RAM banks - the most common controller on
// Game Boy Color-era carts. Unlike MBC1, bank 0 is a genuinely valid
// selection for the switchable region (MBC1 silently bumps it to 1;
// MBC5 does not).
type mbc5 struct {
	rom []byte
	ram []byte

	ramEnabled bool
	romBank    uint16 // 9 bits
	ramBank    uint8  // 4 bits (rumble carts use bit 3 for the motor instead; ignored here)
	hasRumble  bool
}

func newMBC5(cart *Cartridge) *mbc5 {
	return &mbc5{
		rom:       cart.ROM,
		ram:       make([]byte, max(cart.RAMSize, 0)),
		romBank:   1,
		hasRumble: cart.Type >= 0x1C && cart.Type <= 0x1E,
	}
}

func (m *mbc5) WriteROM(addr uint16, v byte) {
	switch {
	case addr <= 0x1FFF:
		m.ramEnabled = v&0x0F == 0x0A
	case addr <= 0x2FFF:
		m.romBank = (m.romBank &^ 0x00FF) | uint16(v)
	case addr <= 0x3FFF:
		m.romBank = (m.romBank &^ 0x0100) | (uint16(v&0x01) << 8)
	case addr <= 0x5FFF:
		bank := v & 0x0F
		if m.hasRumble {
			bank &^= 0x08 // bit 3 is the rumble motor line, not a RAM bank bit
		}
		m.ramBank = bank
	}
}

func (m *mbc5) ReadROM(addr uint16) byte {
	var offset int
	if addr < 0x4000 {
		offset = int(addr)
	} else {
		offset = int(m.romBank)*0x4000 + int(addr-0x4000)
	}
	if offset >= len(m.rom) {
		return 0xFF
	}
	return m.rom[offset]
}

func (m *mbc5) ramOffset(addr uint16) int {
	return int(m.ramBank)*0x2000 + int(addr-0xA000)
}

func (m *mbc5) ReadRAM(addr uint16) byte {
	if !m.ramEnabled {
		return 0xFF
	}
	idx := m.ramOffset(addr)
	if idx < 0 || idx >= len(m.ram) {
		return 0xFF
	}
	return m.ram[idx]
}

func (m *mbc5) WriteRAM(addr uint16, v byte) {
	if !m.ramEnabled {
		return
	}
	idx := m.ramOffset(addr)
	if idx < 0 || idx >= len(m.ram) {
		return
	}
	m.ram[idx] = v
}

func (m *mbc5) Snapshot() MBCSnapshot {
	ram := make([]byte, len(m.ram))
	copy(ram, m.ram)
	return MBCSnapshot{
		RAM: ram,
		// MBCSnapshot's bank fields are only a byte wide (every other MBC
		// here has an 8-bit-or-smaller bank number); MBC5's 9-bit ROM bank
		// needs its own field.
		ROMBank:    uint8(m.romBank),
		ROMBankHi:  uint8(m.romBank >> 8),
		RAMBank:    m.ramBank,
		RAMEnabled: m.ramEnabled,
	}
}

func (m *mbc5) Restore(s MBCSnapshot) {
	copy(m.ram, s.RAM)
	m.romBank = uint16(s.ROMBankHi)<<8 | uint16(s.ROMBank)
	m.ramBank = s.RAMBank
	m.ramEnabled = s.RAMEnabled
}
