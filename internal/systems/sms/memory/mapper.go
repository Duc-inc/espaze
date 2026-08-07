package memory

const romBankSize = 0x4000 // 16KB

// mapper implements the "Sega mapper", the de facto standard on nearly
// every SMS cartridge: three switchable 16KB slots covering $0000-$BFFF,
// each independently pointed at any 16KB bank of the ROM via a register
// at $FFFD/$FFFE/$FFFF. The first 1KB ($0000-$03FF) is a real hardware
// exception - it always reads from ROM bank 0 regardless of what slot 0
// is currently switched to, so the reset/interrupt vectors there never
// move even mid-game.
type mapper struct {
	rom []byte
	ram [0x2000]byte // optional on-cartridge RAM (rarely used by real carts, but cheap to always provide)

	slot0, slot1, slot2 byte
	ramMapped           bool
}

func newMapper(rom []byte) *mapper {
	return &mapper{rom: rom, slot1: 1, slot2: 2}
}

func (m *mapper) bankCount() int {
	if len(m.rom) == 0 {
		return 1
	}
	return len(m.rom) / romBankSize
}

func (m *mapper) romOffset(bank byte, addr uint16) int {
	count := m.bankCount()
	if count == 0 {
		return 0
	}
	return (int(bank)%count)*romBankSize + int(addr)
}

// ReadROM implements $0000-$BFFF.
func (m *mapper) ReadROM(addr uint16) byte {
	if addr < 0x0400 {
		return m.readROMByte(0, addr)
	}
	switch {
	case addr < 0x4000:
		return m.readROMByte(m.slot0, addr-0x0000)
	case addr < 0x8000:
		return m.readROMByte(m.slot1, addr-0x4000)
	default:
		if m.ramMapped {
			return m.ram[addr-0x8000]
		}
		return m.readROMByte(m.slot2, addr-0x8000)
	}
}

func (m *mapper) readROMByte(bank byte, addr uint16) byte {
	offset := m.romOffset(bank, addr)
	if offset < 0 || offset >= len(m.rom) {
		return 0xFF
	}
	return m.rom[offset]
}

// WriteROM implements writes into $8000-$BFFF, which only matters when
// cartridge RAM is currently mapped there.
func (m *mapper) WriteROM(addr uint16, v byte) {
	if addr >= 0x8000 && m.ramMapped {
		m.ram[addr-0x8000] = v
	}
}

// WriteControl handles a write to $FFFC-$FFFF - the mapper's own
// registers, snooped out of what's otherwise a normal RAM write (see
// bus.go, which also stores the byte into the RAM mirror underneath).
func (m *mapper) WriteControl(addr uint16, v byte) {
	switch addr {
	case 0xFFFC:
		m.ramMapped = v&0x08 != 0
	case 0xFFFD:
		m.slot0 = v
	case 0xFFFE:
		m.slot1 = v
	case 0xFFFF:
		m.slot2 = v
	}
}
