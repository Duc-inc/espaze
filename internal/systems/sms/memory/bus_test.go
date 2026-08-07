package memory

import (
	"testing"

	"github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/systems/sms/psg"
	"github.com/Duc-inc/espaze/internal/systems/sms/vdp"
)

func newTestROM(banks int) []byte {
	rom := make([]byte, banks*romBankSize)
	for bank := 0; bank < banks; bank++ {
		rom[bank*romBankSize] = byte(bank) // tag each bank's first byte
	}
	return rom
}

func newTestBus(rom []byte) *Bus {
	return New(rom, vdp.New(), psg.New())
}

func TestSlot1BankSwitching(t *testing.T) {
	b := newTestBus(newTestROM(5))

	b.Write(0xFFFE, 3) // slot 1 ($4000-$7FFF) -> bank 3
	if got := b.Read(0x4000); got != 3 {
		t.Fatalf("bank tag at $4000 = %#02x, want 3", got)
	}

	b.Write(0xFFFE, 4)
	if got := b.Read(0x4000); got != 4 {
		t.Fatalf("bank tag at $4000 after switch = %#02x, want 4", got)
	}
}

func TestFirst1KBAlwaysReadsBank0(t *testing.T) {
	rom := newTestROM(3)
	rom[2*romBankSize+0x400] = 0xAA // bank 2's byte at the same offset $0400 reads at
	b := newTestBus(rom)
	b.Write(0xFFFD, 2) // slot 0 -> bank 2, but $0000-$03FF should stay bank 0

	if got := b.Read(0x0000); got != 0 {
		t.Fatalf("$0000 = %#02x, want 0 (bank 0's tag, unaffected by slot 0's switch)", got)
	}
	if got := b.Read(0x0400); got != 0xAA {
		t.Fatalf("$0400 = %#02x, want 0xAA (bank 2's byte, slot 0 applies past the first 1KB)", got)
	}
}

func TestMapperControlWriteAlsoUpdatesRAMMirror(t *testing.T) {
	b := newTestBus(newTestROM(2))
	b.Write(0xFFFE, 1) // also lands in the RAM mirror underneath

	if got := b.Read(0xFFFE); got != 1 {
		t.Fatalf("RAM mirror at $FFFE = %#02x, want 1 (mapper writes are snooped, not exclusive)", got)
	}
}

func TestRAMMirroring(t *testing.T) {
	b := newTestBus(newTestROM(1))
	b.Write(0xC010, 0x42)
	if got := b.Read(0xE010); got != 0x42 {
		t.Fatalf("echo of $C010 = %#02x, want 0x42", got)
	}
}

func TestJoypadPortActiveLow(t *testing.T) {
	b := newTestBus(newTestROM(1))
	b.SetButtons(input.State{}.With(Up, true).With(Button1, true))

	got := b.In(0xDC)
	if got&(1<<0) != 0 {
		t.Fatal("Up should read as 0 (active-low) while held")
	}
	if got&(1<<4) != 0 {
		t.Fatal("Button1 should read as 0 (active-low) while held")
	}
	if got&(1<<1) == 0 {
		t.Fatal("Down should read as 1 (released)")
	}
}

func TestIOPortsRouteToVDPAndPSG(t *testing.T) {
	b := newTestBus(newTestROM(1))

	b.Out(0xBF, 0x00) // VDP control port, first latch byte (address low)
	b.Out(0xBF, 0x40) // second byte: code 1 (VRAM write setup, addr 0)
	b.Out(0xBE, 0xAB) // VDP data port write

	// Re-setup a read at the same address to confirm the write landed.
	b.Out(0xBF, 0x00)
	b.Out(0xBF, 0x00) // code 0: VRAM read setup, addr 0
	if got := b.In(0xBE); got != 0xAB {
		t.Fatalf("VRAM[0] via IO ports = %#02x, want 0xAB", got)
	}
}
