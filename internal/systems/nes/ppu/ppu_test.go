package ppu

import "testing"

// testCart is a flat, fully-writable CHR space with a fixed mirroring
// mode, enough to drive the PPU in isolation.
type testCart struct {
	chr    [0x2000]byte
	mirror MirrorMode
}

func (c *testCart) ReadCHR(addr uint16) byte     { return c.chr[addr&0x1FFF] }
func (c *testCart) WriteCHR(addr uint16, v byte) { c.chr[addr&0x1FFF] = v }
func (c *testCart) Mirroring() MirrorMode        { return c.mirror }

func newTestPPU() (*PPU, *testCart) {
	cart := &testCart{mirror: MirrorHorizontal}
	return New(cart), cart
}

func TestPaletteMirroring(t *testing.T) {
	var pal paletteRAM
	pal.write(0x00, 0x0F)
	if got := pal.read(0x10); got != 0x0F {
		t.Fatalf("$3F10 = %#02x, want mirror of $3F00 (0x0F)", got)
	}
	pal.write(0x0C, 0x2A)
	if got := pal.read(0x1C); got != 0x2A {
		t.Fatalf("$3F1C = %#02x, want mirror of $3F0C (0x2A)", got)
	}
	// Non-mirrored sprite slots keep their own value.
	pal.write(0x11, 0x05)
	if got := pal.read(0x11); got != 0x05 {
		t.Fatalf("$3F11 = %#02x, want 0x05 (not mirrored)", got)
	}
}

func TestNametableMirroringHorizontal(t *testing.T) {
	// $2000/$2400 share a bank; $2800/$2C00 share the other.
	if nametableBank(MirrorHorizontal, 0) != nametableBank(MirrorHorizontal, 1) {
		t.Fatal("horizontal mirroring: NT0 and NT1 should share a bank")
	}
	if nametableBank(MirrorHorizontal, 0) == nametableBank(MirrorHorizontal, 2) {
		t.Fatal("horizontal mirroring: NT0 and NT2 should NOT share a bank")
	}
}

func TestNametableMirroringVertical(t *testing.T) {
	// $2000/$2800 share a bank; $2400/$2C00 share the other.
	if nametableBank(MirrorVertical, 0) != nametableBank(MirrorVertical, 2) {
		t.Fatal("vertical mirroring: NT0 and NT2 should share a bank")
	}
	if nametableBank(MirrorVertical, 0) == nametableBank(MirrorVertical, 1) {
		t.Fatal("vertical mirroring: NT0 and NT1 should NOT share a bank")
	}
}

func TestPPUSTATUSReadClearsVBlankAndWriteToggle(t *testing.T) {
	p, _ := newTestPPU()
	p.status = statusVBlank
	p.scroll.write = true

	v := p.ReadRegister(0x2002)
	if v&statusVBlank == 0 {
		t.Fatal("expected the read value to still report VBlank was set")
	}
	if p.status&statusVBlank != 0 {
		t.Fatal("expected VBlank to be cleared after reading PPUSTATUS")
	}
	if p.scroll.write {
		t.Fatal("expected the write toggle to reset after reading PPUSTATUS")
	}
}

func TestPPUADDRThenPPUDATAWriteRoundTrips(t *testing.T) {
	p, _ := newTestPPU()

	p.WriteRegister(0x2006, 0x20) // high byte -> nametable space
	p.WriteRegister(0x2006, 0x05) // low byte -> $2005
	p.WriteRegister(0x2007, 0x42)

	if got := p.nametables[nametableBank(p.mirroring(), 0)][0x05]; got != 0x42 {
		t.Fatalf("nametable[0x05] = %#02x, want 0x42", got)
	}
}

func TestPPUDATAReadIsBufferedExceptForPalette(t *testing.T) {
	p, cart := newTestPPU()
	cart.chr[0x0010] = 0x77

	p.WriteRegister(0x2006, 0x00)
	p.WriteRegister(0x2006, 0x10)
	first := p.ReadRegister(0x2007) // primes the buffer, returns stale (0) data
	if first != 0 {
		t.Fatalf("first buffered read = %#02x, want 0x00 (buffer was empty)", first)
	}
	second := p.ReadRegister(0x2007)
	if second != 0x77 {
		t.Fatalf("second buffered read = %#02x, want 0x77", second)
	}
}

func TestVRAMIncrementRespectsPPUCTRL(t *testing.T) {
	p, _ := newTestPPU()
	p.WriteRegister(0x2000, 0x00) // increment by 1
	p.WriteRegister(0x2006, 0x20)
	p.WriteRegister(0x2006, 0x00)
	p.WriteRegister(0x2007, 0xAA)
	if p.scroll.v != 0x2001 {
		t.Fatalf("v after +1 increment = %#04x, want 0x2001", p.scroll.v)
	}

	p.WriteRegister(0x2000, ctrlVRAMIncrement) // increment by 32
	p.WriteRegister(0x2007, 0xBB)
	if p.scroll.v != 0x2021 {
		t.Fatalf("v after +32 increment = %#04x, want 0x2021", p.scroll.v)
	}
}

func TestNMIFiresAtStartOfVBlankWhenEnabled(t *testing.T) {
	p, _ := newTestPPU()
	p.WriteRegister(0x2000, ctrlNMIEnable)

	// tick() examines the (scanline, dot) pair *before* advancing, so
	// reaching pair (241, 1) takes one more Step call than the raw dot
	// count would suggest.
	dotsUntilVBlank := vblankStartScanline*dotsPerScanline + 2
	nmi := p.Step(dotsUntilVBlank)
	if !nmi {
		t.Fatal("expected NMI to fire at the start of vblank")
	}
	if p.status&statusVBlank == 0 {
		t.Fatal("expected VBlank flag set")
	}
}

func TestNoNMIWhenDisabled(t *testing.T) {
	p, _ := newTestPPU()
	dotsUntilVBlank := vblankStartScanline*dotsPerScanline + 1
	if p.Step(dotsUntilVBlank) {
		t.Fatal("did not expect NMI: PPUCTRL NMI-enable bit was never set")
	}
}

func TestBackgroundRendersATile(t *testing.T) {
	p, cart := newTestPPU()

	// Tile #1's pattern data: a solid color-index-3 tile (both bit
	// planes all 1s).
	for i := 0; i < 8; i++ {
		cart.chr[0x0010+i] = 0xFF
		cart.chr[0x0018+i] = 0xFF
	}
	// Nametable entry (0,0) points at tile 1; its attribute byte selects
	// background sub-palette 0.
	p.nametables[nametableBank(p.mirroring(), 0)][0] = 1
	// Palette: sub-palette 0, color index 3 -> palette color 0x30.
	p.palette.write(0x03, 0x30)

	// maskShowBGLeft8 too, or the leftmost 8 pixels get clipped to the
	// universal background color regardless of what's underneath.
	p.WriteRegister(0x2001, maskShowBG|maskShowBGLeft8)
	p.renderScanline(0)

	got := p.frame.Pixels[0:3]
	want := masterPalette[0x30]
	if got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("pixel (0,0) = %v, want %v", got, want)
	}
}
