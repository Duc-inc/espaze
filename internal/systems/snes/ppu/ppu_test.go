package ppu

import "testing"

func writeVRAMWord(p *PPU, addr uint16, word uint16) {
	p.WriteVRAMAddrLow(byte(addr))
	p.WriteVRAMAddrHigh(byte(addr >> 8))
	p.WriteVRAMDataLow(byte(word))
	p.WriteVRAMDataHigh(byte(word >> 8))
}

func TestVRAMWriteAutoIncrements(t *testing.T) {
	p := New()
	writeVRAMWord(p, 0, 0x1111)
	p.WriteVRAMDataLow(0x22)
	p.WriteVRAMDataHigh(0x22)
	if p.vram[1] != 0x2222 {
		t.Fatalf("VRAM[1] = %#04x, want 0x2222 (address should have auto-incremented)", p.vram[1])
	}
}

func TestCGRAMWriteAndResolve(t *testing.T) {
	p := New()
	p.WriteCGRAMAddr(5)
	p.WriteCGRAMData(0x1F) // low byte: red=0x1F
	p.WriteCGRAMData(0x00) // high byte

	r, _, _ := p.resolveColor(5)
	if r == 0 {
		t.Fatal("expected a non-zero red channel after writing red=0x1F into CGRAM index 5")
	}
}

func TestBackgroundTile4BppDecode(t *testing.T) {
	p := New()
	p.WriteBGControl(0, 0x01) // 4bpp
	p.WriteMainScreen(0x01)   // BG0 enabled

	// Name table entry (0,0): tile 1, palette group 2.
	writeVRAMWord(p, 0, 1|2<<10)

	// Tile 1: row0 planes0/1 word at tile*16+0, planes2/3 at tile*16+8.
	writeVRAMWord(p, 16, 0x0080) // plane0 bit7 set at x=0
	writeVRAMWord(p, 24, 0x8080) // plane2,plane3 bit7 set at x=0

	idx, opaque := p.backgroundPixel(0, 0, 0)
	if !opaque {
		t.Fatal("expected an opaque pixel")
	}
	want := uint16(2)*16 + 0x0D // colorBits = 1|0<<1|1<<2|1<<3 = 0b1101 = 13
	if idx != want {
		t.Fatalf("palette index = %d, want %d", idx, want)
	}
}

func TestSpriteReadDecodesAttributes(t *testing.T) {
	p := New()
	p.WriteOAMByte(0, 10)   // X
	p.WriteOAMByte(1, 20)   // Y
	p.WriteOAMByte(2, 5)    // tile
	p.WriteOAMByte(3, 0xC3) // palette=3, flipH, flipV

	s := p.readSprite(0)
	if s.x != 10 || s.y != 20 || s.tile != 5 {
		t.Fatalf("sprite = %+v, want x=10,y=20,tile=5", s)
	}
	if s.palette != 3 || !s.flipH || !s.flipV {
		t.Fatalf("sprite attrs = %+v, want palette=3,flipH=true,flipV=true", s)
	}
}

func TestVBlankIRQFiresAfterVisibleLines(t *testing.T) {
	p := New()
	irq := byte(0)
	for i := 0; i < totalScanlines*cyclesPerLine; i++ {
		irq |= p.Step(1)
	}
	if irq&IRQVBlank == 0 {
		t.Fatal("expected a VBlank IRQ somewhere in a full frame")
	}
}
