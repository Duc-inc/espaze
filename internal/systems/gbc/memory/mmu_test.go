package memory

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/systems/gameboy/apu"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/joypad"
	dmgmem "github.com/Duc-inc/espaze/internal/systems/gameboy/memory"
	"github.com/Duc-inc/espaze/internal/systems/gameboy/timer"
	"github.com/Duc-inc/espaze/internal/systems/gbc/ppu"
)

type stubMBC struct{ rom, ram [0x100]byte }

func (m *stubMBC) ReadROM(addr uint16) byte     { return m.rom[addr%0x100] }
func (m *stubMBC) WriteROM(addr uint16, v byte) {}
func (m *stubMBC) ReadRAM(addr uint16) byte     { return m.ram[addr%0x100] }
func (m *stubMBC) WriteRAM(addr uint16, v byte) { m.ram[addr%0x100] = v }
func (m *stubMBC) Snapshot() dmgmem.MBCSnapshot { return dmgmem.MBCSnapshot{} }
func (m *stubMBC) Restore(s dmgmem.MBCSnapshot) {}

func newTestMMU() *MMU {
	return New(&stubMBC{}, ppu.New(), timer.New(), joypad.New(), apu.New())
}

func TestSVBKSwitchesHighWRAMBank(t *testing.T) {
	m := newTestMMU()

	m.Write(0xFF70, 0x02) // select bank 2
	m.Write(0xD000, 0xAA)
	m.Write(0xFF70, 0x03) // select bank 3
	m.Write(0xD000, 0xBB)

	m.Write(0xFF70, 0x02)
	if got := m.Read(0xD000); got != 0xAA {
		t.Fatalf("bank 2 = %#02x, want 0xAA", got)
	}
	m.Write(0xFF70, 0x03)
	if got := m.Read(0xD000); got != 0xBB {
		t.Fatalf("bank 3 = %#02x, want 0xBB", got)
	}
}

func TestSVBKZeroBehavesAsBankOne(t *testing.T) {
	m := newTestMMU()
	m.Write(0xFF70, 0x01)
	m.Write(0xD000, 0x42)

	m.Write(0xFF70, 0x00)
	if got := m.Read(0xD000); got != 0x42 {
		t.Fatalf("bank 0 read = %#02x, want 0x42 (bank 0 aliases bank 1)", got)
	}
}

func TestEchoRAMMirrorsWorkRAM(t *testing.T) {
	m := newTestMMU()
	m.Write(0xC010, 0x11)
	m.Write(0xD010, 0x22)

	if got := m.Read(0xE010); got != 0x11 {
		t.Fatalf("echo of $C010 = %#02x, want 0x11", got)
	}
	if got := m.Read(0xF010); got != 0x22 {
		t.Fatalf("echo of $D010 = %#02x, want 0x22", got)
	}
}

func TestKEY1TogglesDoubleSpeedOnArmedWrite(t *testing.T) {
	m := newTestMMU()
	if m.DoubleSpeed() {
		t.Fatal("expected single speed initially")
	}

	m.Write(0xFF4D, 0x01) // armed
	if !m.DoubleSpeed() {
		t.Fatal("expected double speed after an armed KEY1 write")
	}
	if m.Read(0xFF4D)&0x80 == 0 {
		t.Fatal("expected KEY1 bit7 to report double speed")
	}

	m.Write(0xFF4D, 0x01)
	if m.DoubleSpeed() {
		t.Fatal("expected a second armed write to switch back to single speed")
	}
}

func TestHDMATransferCopiesIntoVRAM(t *testing.T) {
	m := newTestMMU()
	for i := 0; i < 0x20; i++ {
		m.Write(0xC000+uint16(i), byte(i+1))
	}

	m.Write(0xFF51, 0xC0) // source high = $C0
	m.Write(0xFF52, 0x00) // source low
	m.Write(0xFF53, 0x00) // dest high (within $8000-$9FFF, added to $8000)
	m.Write(0xFF54, 0x00) // dest low
	m.Write(0xFF55, 0x01) // length = (1+1)*0x10 = 0x20 bytes, general-purpose mode

	for i := 0; i < 0x20; i++ {
		if got := m.Read(0x8000 + uint16(i)); got != byte(i+1) {
			t.Fatalf("VRAM[%d] = %#02x, want %#02x", i, got, byte(i+1))
		}
	}
}
