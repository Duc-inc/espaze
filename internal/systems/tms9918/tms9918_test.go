package tms9918

import "testing"

func writeVRAM(t *TMS9918, addr uint16, v byte) {
	t.WriteControl(byte(addr))
	t.WriteControl(byte(addr>>8)&0x3F | 0x40)
	t.WriteData(v)
}

func setRegister(t *TMS9918, reg byte, v byte) {
	t.WriteControl(v)
	t.WriteControl(0x80 | reg)
}

func TestControlPortTwoByteLatchSetsRegister(t *testing.T) {
	tms := New()
	setRegister(tms, 1, 0x40) // display enable
	if tms.regs[1] != 0x40 {
		t.Fatalf("R1 = %#02x, want 0x40", tms.regs[1])
	}
}

func TestVRAMWriteAndReadRoundTrip(t *testing.T) {
	tms := New()
	writeVRAM(tms, 0x0100, 0xAB)

	// Setting the address in read mode (bit6 clear) primes the read
	// buffer immediately, so a single ReadData call returns it.
	tms.WriteControl(0x00)
	tms.WriteControl(0x01)
	if v := tms.ReadData(); v != 0xAB {
		t.Fatalf("ReadData = %#02x, want 0xAB", v)
	}
}

func TestBackgroundPixelUsesPerTileColor(t *testing.T) {
	tms := New()
	setRegister(tms, 1, 0x40) // display enable

	writeVRAM(tms, nameTableBase, 5)           // tile (0,0) uses pattern 5
	writeVRAM(tms, patternTableBase+5*8, 0x80) // pattern 5 row0: leftmost pixel set
	writeVRAM(tms, colorTableBase, 0x2<<4|0x1) // fg=2 (green), bg=1 (black)

	if idx := tms.backgroundPixel(0, 0); idx != 2 {
		t.Fatalf("backgroundPixel(0,0) = %d, want 2 (foreground)", idx)
	}
	if idx := tms.backgroundPixel(7, 0); idx != 1 {
		t.Fatalf("backgroundPixel(7,0) = %d, want 1 (background, pattern bit clear)", idx)
	}
}

func TestSpriteReadAppliesYOffset(t *testing.T) {
	tms := New()
	writeVRAM(tms, spriteAttrBase, 9)    // Y (stored as actual-1)
	writeVRAM(tms, spriteAttrBase+1, 20) // X
	writeVRAM(tms, spriteAttrBase+2, 3)  // tile
	writeVRAM(tms, spriteAttrBase+3, 6)  // color

	s := tms.readSprite(0)
	if s.y != 10 || s.x != 20 || s.tile != 3 || s.color != 6 {
		t.Fatalf("sprite = %+v, want y=10,x=20,tile=3,color=6", s)
	}
}

func TestVBlankIRQFiresWhenEnabled(t *testing.T) {
	tms := New()
	setRegister(tms, 1, 0x60) // display + IRQ enable

	irq := byte(0)
	for i := 0; i < totalScanlines*cyclesPerLine; i++ {
		irq |= tms.Step(1)
	}
	if irq&IRQVBlank == 0 {
		t.Fatal("expected a VBlank IRQ somewhere in a full frame")
	}
}
