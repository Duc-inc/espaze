package powerpc

import "testing"

func mtsprInstr(rS, spr uint32) uint32 {
	return uint32(31)<<26 | rS<<21 | (spr&0x1F)<<16 | ((spr>>5)&0x1F)<<11 | 467<<1
}

func mfsprInstr(rD, spr uint32) uint32 {
	return uint32(31)<<26 | rD<<21 | (spr&0x1F)<<16 | ((spr>>5)&0x1F)<<11 | 339<<1
}

func TestTranslateWithNoBATsConfiguredIsIdentity(t *testing.T) {
	c, _ := newTestCPU(nil)
	if got := c.translate(0x12345678); got != 0x12345678 {
		t.Fatalf("got %#08x, want identity passthrough", got)
	}
}

func TestSetBATConfiguresTranslationDirectly(t *testing.T) {
	c, _ := newTestCPU(nil)
	c.SetBAT(4, 0x80000000, 24*1024*1024, 0)

	if got := c.translate(0x80000000); got != 0 {
		t.Fatalf("translate(0x80000000) = %#08x, want 0", got)
	}
	if got := c.translate(0x81200000); got != 0x01200000 {
		t.Fatalf("translate(0x81200000) = %#08x, want 0x01200000", got)
	}
	if got := c.translate(0x80000000 + 24*1024*1024); got != 0x80000000+24*1024*1024 {
		t.Fatalf("translate just past the block end = %#08x, want unchanged", got)
	}
}

func TestSetBATOutOfRangeIndexIsIgnored(t *testing.T) {
	c, _ := newTestCPU(nil)
	c.SetBAT(99, 0x80000000, 1024*1024, 0) // must not panic
	if got := c.translate(0x80000000); got != 0x80000000 {
		t.Fatalf("got %#08x, want unchanged (out-of-range index ignored)", got)
	}
}

func TestMtsprConfiguresBATAndTranslateMaps(t *testing.T) {
	// DBAT0: map effective 0xC0000000, 24MB, onto physical 0x00000000
	// - this project's own approximation of the real GameCube's
	// cached (0x80000000) vs uncached (0xC0000000) MEM1 aliasing.
	const blockSize = 24 * 1024 * 1024
	bl := uint32(blockSize/(128*1024)) - 1
	batu := (uint32(0xC0000000) >> 17 << 17) | bl<<2 | 0x3 // BEPI | BL | Vs|Vp
	batl := uint32(0)                                      // BRPN=0 -> physical base 0

	c, _ := newTestCPU([]uint32{
		mtsprInstr(1, 536), // DBAT0U (SPR 536)
		mtsprInstr(2, 537), // DBAT0L (SPR 537)
	})
	c.SetGPR(1, batu)
	c.SetGPR(2, batl)
	c.Step()
	c.Step()

	if got := c.translate(0xC0000000); got != 0 {
		t.Fatalf("translate(0xC0000000) = %#08x, want 0", got)
	}
	if got := c.translate(0xC0000010); got != 0x10 {
		t.Fatalf("translate(0xC0000010) = %#08x, want 0x10", got)
	}
	if got := c.translate(0xC0000000 + blockSize); got != 0xC0000000+blockSize {
		t.Fatalf("translate just past the block end = %#08x, want unchanged (outside the BAT)", got)
	}
}

func TestMfsprReadsBackConfiguredBAT(t *testing.T) {
	c, _ := newTestCPU([]uint32{
		mtsprInstr(1, 528), // IBAT0U
		mfsprInstr(2, 528), // read it back
	})
	c.SetGPR(1, 0x12345678&^0x3|0x3) // keep Vs|Vp set so it round-trips through decode/encode
	c.Step()
	c.Step()

	want := encodeBATUpper(bat{
		effective: (0x12345678 &^ 0x3) >> 17,
		length:    (0x12345678 >> 2) & 0x7FF,
		valid:     true,
	})
	if c.GPR(2) != want {
		t.Fatalf("GPR2 = %#08x, want %#08x", c.GPR(2), want)
	}
}

func TestLoadStoreGoThroughBATTranslation(t *testing.T) {
	// Map effective 0x10000000 onto physical 0x00000000, then store
	// through the effective address and read back through the
	// physical one (and vice versa) to prove the load/store opcode
	// paths actually call translate, not just the translate function
	// in isolation.
	const blockSize = 1024 * 1024
	bl := uint32(blockSize/(128*1024)) - 1
	batu := (uint32(0x10000000) >> 17 << 17) | bl<<2 | 0x3
	batl := uint32(0)

	program := []uint32{
		mtsprInstr(1, 536), // DBAT0U
		mtsprInstr(2, 537), // DBAT0L
		// addi r3,r0,0x55 ; stw r3,0x50(r0) using effective base r4=0x10000000
		uint32(14)<<26 | 3<<21 | 0<<16 | 0x55,
		uint32(36)<<26 | 3<<21 | 4<<16 | 0x50, // stw r3, 0x50(r4)
		uint32(32)<<26 | 5<<21 | 0<<16 | 0x50, // lwz r5, 0x50(r0) - physical, no translation applies (r0-based, still goes through translate but 0x50 isn't in any BAT)
	}
	c, bus := newTestCPU(program)
	c.SetGPR(1, batu)
	c.SetGPR(2, batl)
	c.SetGPR(4, 0x10000000)
	c.Step() // mtspr
	c.Step() // mtspr
	c.Step() // addi
	c.Step() // stw through effective 0x10000050 -> physical 0x50
	c.Step() // lwz from physical 0x50 directly

	if c.GPR(5) != 0x55 {
		t.Fatalf("GPR5 = %#x, want 0x55 (value written through the translated effective address)", c.GPR(5))
	}
	if bus.Read32(0x50) != 0x55 {
		t.Fatalf("bus[0x50] = %#x, want 0x55", bus.Read32(0x50))
	}
}
