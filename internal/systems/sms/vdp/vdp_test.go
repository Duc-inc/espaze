package vdp

import "testing"

func writeVRAMSetup(v *VDP, addr uint16) {
	v.WriteControl(byte(addr))
	v.WriteControl(byte(addr>>8) | 0x40) // code 1 = VRAM write
}

func writeRegister(v *VDP, reg byte, val byte) {
	v.WriteControl(val)
	v.WriteControl(0x80 | reg)
}

func writeCRAMSetup(v *VDP, addr byte) {
	v.WriteControl(addr)
	v.WriteControl(0xC0)
}

func TestPaletteExpansion(t *testing.T) {
	var c cram
	c.write(0, 0x3F) // both bits of every channel set: max R, G, B
	r, g, b := c.rgb(0)
	if r != 255 || g != 255 || b != 255 {
		t.Fatalf("rgb = %d,%d,%d, want 255,255,255", r, g, b)
	}
}

func TestControlPortRegisterWrite(t *testing.T) {
	v := New()
	writeRegister(v, 1, 0x40) // R1 = display enable

	if v.regs[1] != 0x40 {
		t.Fatalf("R1 = %#02x, want 0x40", v.regs[1])
	}
	if !v.displayEnabled() {
		t.Fatal("expected display enabled")
	}
}

func TestVRAMWriteAndReadRoundTrip(t *testing.T) {
	v := New()
	writeVRAMSetup(v, 0x1000)
	v.WriteData(0xAB)
	v.WriteData(0xCD)

	if v.vram[0x1000] != 0xAB || v.vram[0x1001] != 0xCD {
		t.Fatalf("vram[0x1000:2] = %#02x %#02x, want AB CD", v.vram[0x1000], v.vram[0x1001])
	}

	// Re-setup for read: the read-ahead buffer means the first readData()
	// returns the byte primed at setup time, not the one just written.
	v.WriteControl(0x00)
	v.WriteControl(0x10) // code 0 = read setup, addr 0x1000
	first := v.ReadData()
	if first != 0xAB {
		t.Fatalf("first read = %#02x, want 0xAB", first)
	}
	second := v.ReadData()
	if second != 0xCD {
		t.Fatalf("second read = %#02x, want 0xCD", second)
	}
}

func TestCRAMWrite(t *testing.T) {
	v := New()
	writeCRAMSetup(v, 5)
	v.WriteData(0x2A)

	if v.cram.read(5) != 0x2A {
		t.Fatalf("cram[5] = %#02x, want 0x2A", v.cram.read(5))
	}
}

func TestStatusReadClearsFlagsAndLatch(t *testing.T) {
	v := New()
	v.status = statusVBlank | statusSpriteCollide
	v.ctrlLatched = true

	got := v.ReadStatus()
	if got&statusVBlank == 0 {
		t.Fatal("expected the read value to report VBlank")
	}
	if v.status != 0 {
		t.Fatal("expected status to be cleared after reading it")
	}
	if v.ctrlLatched {
		t.Fatal("expected the control port latch to reset")
	}
}

func TestFrameIRQFiresAtStartOfVBlank(t *testing.T) {
	v := New()
	writeRegister(v, 1, reg1FrameIRQ)

	var irq byte
	for line := 0; line <= Height && irq&IRQFrame == 0; line++ {
		irq = v.Step(cyclesPerLine)
	}
	if irq&IRQFrame == 0 {
		t.Fatal("expected the frame IRQ to fire once VBlank starts")
	}
	if v.status&statusVBlank == 0 {
		t.Fatal("expected the VBlank status flag set")
	}
}

func TestBackgroundTileRendersWithPalette(t *testing.T) {
	v := New()
	writeRegister(v, 1, reg1DisplayEnable)

	// Tile 1, solid color index 15 (all 4 bit-planes set) at row 0.
	writeVRAMSetup(v, 32) // tile 1 starts at byte 32 (tile 0 occupies 0-31)
	for i := 0; i < 4; i++ {
		v.WriteData(0xFF)
	}
	// Name table entry (0,0) at the default base ($0000): tile index 1,
	// palette select 0.
	writeVRAMSetup(v, 0)
	v.WriteData(0x01) // low byte of tile index
	v.WriteData(0x00) // high bit + flags

	writeCRAMSetup(v, 15)
	v.WriteData(0x3F) // palette 0, color 15: full white

	v.renderScanline(0)

	if got := v.frame.Pixels[0]; got != 255 {
		t.Fatalf("pixel(0,0) red = %d, want 255", got)
	}
}
