package tia

import "testing"

func TestPlayfieldBitOrderMatchesHardware(t *testing.T) {
	pf := playfield{pf0: 0xF0, pf1: 0x00, pf2: 0x00}
	for i := 0; i < 4; i++ {
		if !pf.bit(i) {
			t.Fatalf("PF0 bit %d should be set (PF0=0xF0)", i)
		}
	}
	if pf.bit(4) {
		t.Fatal("PF1 bit 0 should be clear (PF1=0x00)")
	}
}

func TestPlayfieldReflectsRightHalf(t *testing.T) {
	pf := playfield{pf0: 0x10, reflect: true} // only bit 0 set
	if !pf.pixelAt(0) {
		t.Fatal("left half bit 0 should be set")
	}
	if !pf.pixelAt(159) {
		t.Fatal("reflected right half should mirror bit 0 onto the last pixel")
	}
}

func TestPlayerPixelHonorsReflectAndScale(t *testing.T) {
	p := newPlayer()
	p.grp = 0x80 // only the leftmost pixel set
	p.pos = 10

	if !p.pixelAt(10) {
		t.Fatal("expected the leftmost graphics bit to draw at pos")
	}
	if p.pixelAt(17) {
		t.Fatal("did not expect the rightmost pixel to be set")
	}

	p.reflect = true
	if p.pixelAt(10) {
		t.Fatal("reflected player should no longer draw at pos")
	}
	if !p.pixelAt(17) {
		t.Fatal("reflected player should draw its bit at the far end instead")
	}
}

func TestMovableWidthAndWrap(t *testing.T) {
	m := newMovable()
	m.enabled = true
	m.width = 4
	m.pos = 158

	if !m.pixelAt(159) {
		t.Fatal("expected pixel 159 to be covered")
	}
	if !m.pixelAt(0) {
		t.Fatal("expected the object to wrap around to pixel 0")
	}
	if m.pixelAt(3) {
		t.Fatal("did not expect pixel 3 to be covered by a width-4 object at 158")
	}
}

func TestStepRendersAVisiblePixel(t *testing.T) {
	tia := New()
	tia.WriteRegister(0x09, 0x1E) // COLUBK: a non-black background color

	// Advance past hblank and into the first visible line.
	cyclesToFirstPixel := (firstVisibleLine*colorClocksPerLine + hblankClocks + 1) / 3
	tia.Step(cyclesToFirstPixel)

	r, g, b, _ := colorAt(tia, 0, 0)
	if r == 0 && g == 0 && b == 0 {
		t.Fatal("expected a non-black background pixel to have been rendered")
	}
}

func colorAt(t *TIA, x, y int) (byte, byte, byte, byte) {
	fb := t.FrameBuffer()
	i := (y*fb.Width + x) * 4
	return fb.Pixels[i], fb.Pixels[i+1], fb.Pixels[i+2], fb.Pixels[i+3]
}

func TestWSyncClearsAtNextLine(t *testing.T) {
	tia := New()
	tia.WriteRegister(0x02, 0) // WSYNC
	if !tia.WSyncPending() {
		t.Fatal("expected WSYNC to be pending immediately after the strobe")
	}
	tia.Step(colorClocksPerLine) // one full scanline's worth of CPU cycles (well over 228/3)
	if tia.WSyncPending() {
		t.Fatal("expected WSYNC to clear once the line wrapped")
	}
}

func TestFrameDoneAfterFullRaster(t *testing.T) {
	tia := New()
	cyclesPerFrame := totalScanlines * colorClocksPerLine / 3
	tia.Step(cyclesPerFrame)
	if !tia.FrameDone() {
		t.Fatal("expected FrameDone after a full 262-line raster")
	}
}

func TestCollisionRegisterSetOnOverlap(t *testing.T) {
	tia := New()
	tia.p0.pos = 10
	tia.p0.grp = 0xFF
	tia.p1.pos = 10
	tia.p1.grp = 0xFF

	tia.updateCollisions(true, true, false, false, false, false)
	if tia.cxppmm&0x80 == 0 {
		t.Fatal("expected CXPPMM player0/player1 bit to be set")
	}
}
