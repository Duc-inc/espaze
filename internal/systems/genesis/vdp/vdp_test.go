package vdp

import "testing"

// vdpCommand splits a 6-bit access code and 16-bit VRAM/CRAM/VSRAM
// address into the two 16-bit words the real control port protocol
// expects: CD1-CD0 (the code's low 2 bits) travel in word1 alongside
// address bits 13-0; CD5-CD2 (the code's high 4 bits) travel in word2
// alongside address bits 15-14.
func vdpCommand(code byte, addr uint16) (word1, word2 uint16) {
	word1 = uint16(code&0x03)<<14 | addr&0x3FFF
	word2 = uint16(code&0x3C)<<2 | (addr>>14)&0x03
	return
}

func sendCommand(v *VDP, code byte, addr uint16) {
	w1, w2 := vdpCommand(code, addr)
	v.writeControl(w1)
	v.writeControl(w2)
}

func writeRegister(v *VDP, reg byte, val byte) {
	v.writeControl(0x8000 | uint16(reg)<<8 | uint16(val))
}

func TestRegisterWriteSingleWord(t *testing.T) {
	v := New()
	writeRegister(v, 1, 0x40) // display enable
	if v.regs[1] != 0x40 {
		t.Fatalf("R1 = %#02x, want 0x40", v.regs[1])
	}
	if !v.displayEnabled() {
		t.Fatal("expected display enabled")
	}
}

func TestVRAMWriteAndReadRoundTrip(t *testing.T) {
	v := New()
	sendCommand(v, codeVRAMWrite, 0x1000)
	v.WriteData(0xABCD)

	sendCommand(v, codeVRAMRead, 0x1000)
	got := v.ReadData()
	if got != 0xABCD {
		t.Fatalf("VRAM[0x1000] = %#04x, want 0xABCD", got)
	}
}

func TestCRAMWrite(t *testing.T) {
	v := New()
	sendCommand(v, codeCRAMWrite, 0)
	v.WriteData(0x0EEE) // max R/G/B

	r, g, b := v.palette.rgb(0)
	if r != 255 || g != 255 || b != 255 {
		t.Fatalf("rgb = %d,%d,%d, want 255,255,255", r, g, b)
	}
}

func TestVSRAMWrite(t *testing.T) {
	v := New()
	sendCommand(v, codeVSRAMWrite, 2) // VSRAM word index 1
	v.WriteData(0x0123)

	if v.vsram[1] != 0x0123 {
		t.Fatalf("vsram[1] = %#04x, want 0x0123", v.vsram[1])
	}
}

func TestDMAMemoryToVRAM(t *testing.T) {
	v := New()
	mem := &fakeMemory{}
	mem.data[0x1000/2] = 0x1234
	mem.data[0x1002/2] = 0x5678
	v.SetMemory(mem)

	writeRegister(v, 1, 0x10) // DMA enable
	writeRegister(v, 19, 2)   // DMA length = 2 words
	writeRegister(v, 20, 0)
	writeRegister(v, 21, byte((0x1000>>1)&0xFF))  // source address bits 7-0 (>>1)
	writeRegister(v, 22, byte((0x1000>>9)&0xFF))  // source address bits 15-8
	writeRegister(v, 23, byte((0x1000>>17)&0x7F)) // source address bits 22-16, bit7=0 (68k->VDP mode)
	writeRegister(v, 15, 2)                       // auto-increment by 2

	sendCommand(v, codeVRAMWrite|codeDMABit, 0x2000)

	if v.vramWord(0x2000) != 0x1234 {
		t.Fatalf("VRAM[0x2000] = %#04x, want 0x1234", v.vramWord(0x2000))
	}
	if v.vramWord(0x2002) != 0x5678 {
		t.Fatalf("VRAM[0x2002] = %#04x, want 0x5678", v.vramWord(0x2002))
	}
}

func TestDMAFillTriggersOnNextDataWrite(t *testing.T) {
	v := New()
	writeRegister(v, 1, 0x10) // DMA enable
	writeRegister(v, 19, 4)   // fill 4 bytes
	writeRegister(v, 20, 0)
	writeRegister(v, 15, 1)    // auto-increment by 1
	writeRegister(v, 23, 0x80) // fill mode (bit7 set, bit6 clear)

	// A VRAM-write control command with the DMA bit set arms the fill;
	// the *next* data-port write's high byte becomes the fill value.
	sendCommand(v, codeVRAMWrite|codeDMABit, 0x3000)
	if !v.dmaFillArmed {
		t.Fatal("expected fill DMA to be armed after the control command")
	}

	v.WriteData(0xAB00)
	if v.dmaFillArmed {
		t.Fatal("expected the fill to have completed and cleared the armed flag")
	}
	for i := uint16(0); i < 4; i++ {
		if got := v.vram[0x3000+i]; got != 0xAB {
			t.Fatalf("vram[%#04x] = %#02x, want 0xAB", 0x3000+i, got)
		}
	}
}

func TestBackgroundTileRenders(t *testing.T) {
	v := New()
	writeRegister(v, 1, 0x40) // display enable

	// Tile 1: solid color index 5 (both nibbles of every byte = 5).
	base := uint16(1) * 32
	for i := uint16(0); i < 32; i++ {
		v.vram[base+i] = 0x55
	}
	// Plane A name table (default base $0000): entry (0,0) -> tile 1,
	// palette line 0.
	v.vram[0], v.vram[1] = 0x00, 0x01

	v.palette.write(5, 0x0E00) // palette 0, color 5: full blue

	v.renderScanline(0)
	r, g, b := v.frame.Pixels[0], v.frame.Pixels[1], v.frame.Pixels[2]
	if r != 0 || g != 0 || b != 255 {
		t.Fatalf("pixel(0,0) = %d,%d,%d, want 0,0,255 (blue)", r, g, b)
	}
}

type fakeMemory struct {
	data [0x10000]uint16
}

func (m *fakeMemory) Read16(addr uint32) uint16 { return m.data[(addr&0xFFFF)/2] }
