package memory

import "github.com/Duc-inc/espaze/internal/systems/nes/ppu"

// uxrom implements mapper 2: a switchable 16KB PRG bank at $8000-$BFFF,
// with the cartridge's last 16KB bank fixed at $C000-$FFFF. CHR is
// always RAM on real UxROM boards.
type uxrom struct {
	cart  *Cartridge
	bank  int
	banks int
}

func newUxROM(cart *Cartridge) *uxrom {
	return &uxrom{cart: cart, banks: len(cart.PRG) / prgBankSize}
}

func (m *uxrom) ReadPRG(addr uint16) byte {
	if addr >= 0xC000 {
		offset := (m.banks-1)*prgBankSize + int(addr-0xC000)
		return m.cart.PRG[offset]
	}
	offset := m.bank*prgBankSize + int(addr-0x8000)
	return m.cart.PRG[offset]
}

func (m *uxrom) WritePRG(addr uint16, v byte) {
	m.bank = int(v) % m.banks
}

func (m *uxrom) ReadCHR(addr uint16) byte { return m.cart.CHR[addr] }
func (m *uxrom) WriteCHR(addr uint16, v byte) {
	if m.cart.ChrIsRAM {
		m.cart.CHR[addr] = v
	}
}

func (m *uxrom) Mirroring() ppu.MirrorMode { return m.cart.Mirror }

func (m *uxrom) Snapshot() []byte { return gobEncode(m.bank) }
func (m *uxrom) Restore(data []byte) error {
	return gobDecode(data, &m.bank)
}
