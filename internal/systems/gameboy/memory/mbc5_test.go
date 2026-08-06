package memory

import "testing"

func newTestMBC5Cart(banks int) *Cartridge {
	rom := make([]byte, banks*0x4000)
	for bank := 0; bank < banks; bank++ {
		rom[bank*0x4000] = byte(bank) // tag each bank's first byte with its own number
	}
	return &Cartridge{Type: 0x19, ROM: rom, RAMSize: 8 * 1024}
}

func TestMBC5NineBitBankSwitching(t *testing.T) {
	m := newMBC5(newTestMBC5Cart(300)) // more than MBC1's 5-bit max of 31

	m.WriteROM(0x2000, 0x01) // low 8 bits of the bank number
	m.WriteROM(0x3000, 0x01) // bit 8 -> bank 0x101 = 257

	if got := m.ReadROM(0x4000); got != 1 {
		t.Fatalf("bank 257's tag byte = %#02x, want 0x01", got)
	}
}

func TestMBC5Bank0IsSelectableUnlikeMBC1(t *testing.T) {
	m := newMBC5(newTestMBC5Cart(4))
	m.WriteROM(0x2000, 0x00) // MBC1 would silently bump this to 1; MBC5 must not

	if got := m.ReadROM(0x4000); got != 0 {
		t.Fatalf("bank 0's tag byte = %#02x, want 0x00 (MBC5 allows bank 0 here)", got)
	}
}

func TestMBC5RAMEnableGatesReadsAndWrites(t *testing.T) {
	m := newMBC5(newTestMBC5Cart(2))

	m.WriteRAM(0xA000, 0x42)
	if got := m.ReadRAM(0xA000); got != 0xFF {
		t.Fatalf("read with RAM disabled = %#02x, want 0xFF", got)
	}

	m.WriteROM(0x0000, 0x0A) // enable
	m.WriteRAM(0xA000, 0x42)
	if got := m.ReadRAM(0xA000); got != 0x42 {
		t.Fatalf("read with RAM enabled = %#02x, want 0x42", got)
	}
}

func TestMBC5RumbleBitIsMaskedOutOfRAMBank(t *testing.T) {
	m := newMBC5(newTestMBC5Cart(2))
	m.hasRumble = true
	m.WriteROM(0x0000, 0x0A)

	m.WriteROM(0x4000, 0x0F) // bit3 (rumble motor) set alongside a RAM bank
	if m.ramBank&0x08 != 0 {
		t.Fatalf("ramBank = %#02x, rumble bit should have been masked out", m.ramBank)
	}
}
