package memory

import "github.com/Duc-inc/espaze/internal/systems/nes/ppu"

// nrom implements mapper 0: no bank switching at all. PRG is 16KB
// (mirrored across both halves of $8000-$FFFF) or 32KB (filling it
// directly); CHR is a single fixed 8KB bank, ROM or RAM.
type nrom struct {
	cart *Cartridge
}

func newNROM(cart *Cartridge) *nrom { return &nrom{cart: cart} }

func (m *nrom) ReadPRG(addr uint16) byte {
	offset := int(addr - 0x8000)
	if len(m.cart.PRG) == prgBankSize {
		offset %= prgBankSize
	}
	return m.cart.PRG[offset]
}

func (m *nrom) WritePRG(addr uint16, v byte) {} // no PRG registers on plain NROM

func (m *nrom) ReadCHR(addr uint16) byte { return m.cart.CHR[addr] }
func (m *nrom) WriteCHR(addr uint16, v byte) {
	if m.cart.ChrIsRAM {
		m.cart.CHR[addr] = v
	}
}

func (m *nrom) Mirroring() ppu.MirrorMode { return m.cart.Mirror }

// NROM has no switchable registers, so there's nothing to snapshot.
func (m *nrom) Snapshot() []byte          { return nil }
func (m *nrom) Restore(data []byte) error { return nil }
