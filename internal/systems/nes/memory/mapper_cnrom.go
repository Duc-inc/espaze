package memory

import "github.com/Duc-inc/espaze/internal/systems/nes/ppu"

// cnrom implements mapper 3: fixed PRG (like NROM), with CHR switched
// in fixed 8KB banks by writing anywhere in $8000-$FFFF.
type cnrom struct {
	cart     *Cartridge
	chrBank  int
	chrBanks int
}

func newCNROM(cart *Cartridge) *cnrom {
	banks := len(cart.CHR) / chrBankSize
	if banks == 0 {
		banks = 1
	}
	return &cnrom{cart: cart, chrBanks: banks}
}

func (m *cnrom) ReadPRG(addr uint16) byte {
	offset := int(addr - 0x8000)
	if len(m.cart.PRG) == prgBankSize {
		offset %= prgBankSize
	}
	return m.cart.PRG[offset]
}

func (m *cnrom) WritePRG(addr uint16, v byte) {
	m.chrBank = int(v) % m.chrBanks
}

func (m *cnrom) ReadCHR(addr uint16) byte {
	return m.cart.CHR[m.chrBank*chrBankSize+int(addr)]
}

func (m *cnrom) WriteCHR(addr uint16, v byte) {
	if m.cart.ChrIsRAM {
		m.cart.CHR[m.chrBank*chrBankSize+int(addr)] = v
	}
}

func (m *cnrom) Mirroring() ppu.MirrorMode { return m.cart.Mirror }

func (m *cnrom) Snapshot() []byte { return gobEncode(m.chrBank) }
func (m *cnrom) Restore(data []byte) error {
	return gobDecode(data, &m.chrBank)
}
