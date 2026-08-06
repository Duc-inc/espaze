package memory

import "github.com/Duc-inc/espaze/internal/systems/nes/ppu"

// mmc1 implements mapper 1: bank switching driven by a 5-bit serial
// shift register loaded one bit at a time (LSB first) over 5
// consecutive writes anywhere in $8000-$FFFF, latched into one of four
// internal registers based on which address range the 5th write landed
// in. Used by a large share of early-to-mid NES releases (Zelda,
// Metroid, Mega Man 2...).
type mmc1 struct {
	cart *Cartridge

	shift      byte
	shiftCount int

	control  byte // bit0-1 mirroring, bit2-3 PRG mode, bit4 CHR mode
	chrBank0 byte
	chrBank1 byte
	prgBank  byte

	prgBanks int
}

func newMMC1(cart *Cartridge) *mmc1 {
	return &mmc1{
		cart:     cart,
		control:  0x0C, // PRG mode 3 (fix last bank) is the post-reset default
		prgBanks: len(cart.PRG) / prgBankSize,
	}
}

// WritePRG feeds the shift register; a write with bit7 set resets it
// instead (real hardware does this so a reset mid-sequence can't leave
// a mapper in a half-written state).
func (m *mmc1) WritePRG(addr uint16, v byte) {
	if v&0x80 != 0 {
		m.shift, m.shiftCount = 0, 0
		m.control |= 0x0C
		return
	}

	m.shift |= (v & 1) << uint(m.shiftCount)
	m.shiftCount++
	if m.shiftCount < 5 {
		return
	}

	value := m.shift
	m.shift, m.shiftCount = 0, 0

	switch {
	case addr < 0xA000:
		m.control = value
	case addr < 0xC000:
		m.chrBank0 = value
	case addr < 0xE000:
		m.chrBank1 = value
	default:
		m.prgBank = value & 0x0F
	}
}

func (m *mmc1) ReadPRG(addr uint16) byte {
	offset := int(addr - 0x8000)
	switch (m.control >> 2) & 0x03 {
	case 0, 1: // 32KB mode: ignore the bank register's low bit
		bank := int(m.prgBank &^ 1)
		return m.cart.PRG[(bank*prgBankSize+offset)%len(m.cart.PRG)]
	case 2: // fix first bank at $8000, switch $C000
		if addr < 0xC000 {
			return m.cart.PRG[offset%len(m.cart.PRG)]
		}
		bank := int(m.prgBank) % m.prgBanks
		return m.cart.PRG[bank*prgBankSize+int(addr-0xC000)]
	default: // fix last bank at $C000, switch $8000
		if addr >= 0xC000 {
			bank := m.prgBanks - 1
			return m.cart.PRG[bank*prgBankSize+int(addr-0xC000)]
		}
		bank := int(m.prgBank) % m.prgBanks
		return m.cart.PRG[bank*prgBankSize+offset]
	}
}

func (m *mmc1) chrOffset(addr uint16) int {
	const fourKB = chrBankSize / 2
	if m.control&0x10 == 0 { // 8KB mode: chrBank0 selects, ignoring its low bit
		bank := int(m.chrBank0 &^ 1)
		return bank*fourKB + int(addr)
	}
	if addr < 0x1000 {
		return int(m.chrBank0)*fourKB + int(addr)
	}
	return int(m.chrBank1)*fourKB + int(addr-0x1000)
}

func (m *mmc1) ReadCHR(addr uint16) byte {
	return m.cart.CHR[m.chrOffset(addr)%len(m.cart.CHR)]
}

func (m *mmc1) WriteCHR(addr uint16, v byte) {
	if m.cart.ChrIsRAM {
		m.cart.CHR[m.chrOffset(addr)%len(m.cart.CHR)] = v
	}
}

type mmc1State struct {
	Shift, ShiftCount           byte
	Control, ChrBank0, ChrBank1 byte
	PrgBank                     byte
}

func (m *mmc1) Snapshot() []byte {
	return gobEncode(mmc1State{
		Shift: m.shift, ShiftCount: byte(m.shiftCount),
		Control: m.control, ChrBank0: m.chrBank0, ChrBank1: m.chrBank1, PrgBank: m.prgBank,
	})
}

func (m *mmc1) Restore(data []byte) error {
	var s mmc1State
	if err := gobDecode(data, &s); err != nil {
		return err
	}
	m.shift, m.shiftCount = s.Shift, int(s.ShiftCount)
	m.control, m.chrBank0, m.chrBank1, m.prgBank = s.Control, s.ChrBank0, s.ChrBank1, s.PrgBank
	return nil
}

func (m *mmc1) Mirroring() ppu.MirrorMode {
	switch m.control & 0x03 {
	case 0:
		return ppu.MirrorSingleScreenLow
	case 1:
		return ppu.MirrorSingleScreenHigh
	case 2:
		return ppu.MirrorVertical
	default:
		return ppu.MirrorHorizontal
	}
}
