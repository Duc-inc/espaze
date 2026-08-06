package memory

// MBC abstracts a cartridge's memory bank controller: how CPU reads/writes
// into ROM space (0x0000-0x7FFF) and external RAM space (0xA000-0xBFFF)
// are actually resolved to bytes in the (possibly much larger) ROM/RAM.
type MBC interface {
	ReadROM(addr uint16) byte
	WriteROM(addr uint16, v byte) // writes here are control registers, not data
	ReadRAM(addr uint16) byte
	WriteRAM(addr uint16, v byte)
	Snapshot() MBCSnapshot
	Restore(MBCSnapshot)
}

// MBCSnapshot captures whatever mutable state an MBC has (bank selection,
// RAM contents, enable latch) for save states. Not every field is used by
// every MBC type.
type MBCSnapshot struct {
	RAM         []byte
	ROMBank     uint8
	RAMBank     uint8
	RAMEnabled  bool
	BankingMode uint8
}

// NewMBC picks an implementation based on the cartridge header's type byte.
func NewMBC(cart *Cartridge) MBC {
	switch {
	case cart.Type == 0x00:
		return newMBC0(cart)
	case cart.Type >= 0x01 && cart.Type <= 0x03:
		return newMBC1(cart)
	default:
		// Unsupported controller: fall back to plain ROM mapping so the
		// game at least boots instead of crashing the whole app.
		return newMBC0(cart)
	}
}
