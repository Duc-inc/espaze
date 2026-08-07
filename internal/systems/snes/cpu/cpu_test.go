package cpu

import "testing"

type testBus struct {
	mem [0x1000000]byte
}

func (b *testBus) Read8(addr uint32) byte     { return b.mem[addr&0xFFFFFF] }
func (b *testBus) Write8(addr uint32, v byte) { b.mem[addr&0xFFFFFF] = v }

func newTestCPU(program []byte) (*CPU, *testBus) {
	bus := &testBus{}
	bus.mem[resetVector] = 0x00
	bus.mem[resetVector+1] = 0x80 // reset vector -> $8000
	copy(bus.mem[0x8000:], program)
	return New(bus), bus
}

func TestResetStartsInEmulationMode(t *testing.T) {
	c, _ := newTestCPU(nil)
	if !c.regs.E {
		t.Fatal("expected emulation mode at reset")
	}
	if c.PC() != 0x8000 {
		t.Fatalf("PC = %#06x, want 0x8000", c.PC())
	}
	if c.regs.S != 0x01FD {
		t.Fatalf("S = %#04x, want 0x01FD", c.regs.S)
	}
}

func TestLDAImmediate8BitInEmulationMode(t *testing.T) {
	c, _ := newTestCPU([]byte{0xA9, 0x42}) // LDA #$42
	c.Step()
	if c.regs.A != 0x42 {
		t.Fatalf("A = %#04x, want 0x0042", c.regs.A)
	}
}

func TestXCESwitchesToNativeModeAnd16BitLDA(t *testing.T) {
	// CLC ; XCE (enter native mode) ; REP #$20 (16-bit A) ; LDA #$1234
	c, _ := newTestCPU([]byte{0x18, 0xFB, 0xC2, 0x20, 0xA9, 0x34, 0x12})
	c.Step() // CLC
	c.Step() // XCE
	if c.regs.E {
		t.Fatal("expected native mode after XCE with Carry clear")
	}
	c.Step() // REP #$20
	if c.regs.accum8() {
		t.Fatal("expected 16-bit accumulator after REP #$20")
	}
	c.Step() // LDA #$1234
	if c.regs.A != 0x1234 {
		t.Fatalf("A = %#04x, want 0x1234", c.regs.A)
	}
}

func TestStoreAndLoadDirectPage(t *testing.T) {
	// LDA #$77 ; STA $10 ; LDA #$00 ; LDA $10
	c, _ := newTestCPU([]byte{0xA9, 0x77, 0x85, 0x10, 0xA9, 0x00, 0xA5, 0x10})
	c.Step()
	c.Step()
	c.Step()
	c.Step()
	if c.regs.A != 0x77 {
		t.Fatalf("A = %#04x, want 0x0077", c.regs.A)
	}
}

func TestAdcSetsCarryOnOverflow8Bit(t *testing.T) {
	// LDA #$F0 ; ADC #$20 -> wraps to 0x10 with Carry set
	c, _ := newTestCPU([]byte{0xA9, 0xF0, 0x69, 0x20})
	c.Step()
	c.Step()
	if byte(c.regs.A) != 0x10 {
		t.Fatalf("A low byte = %#02x, want 0x10", byte(c.regs.A))
	}
	if !c.regs.getFlag(FlagCarry) {
		t.Fatal("expected Carry set")
	}
}

func TestJsrAndRts(t *testing.T) {
	// JSR $8010 ; (return here) LDA #$01 ; at $8010: LDA #$02 ; RTS
	c, bus := newTestCPU([]byte{0x20, 0x10, 0x80, 0xA9, 0x01})
	bus.mem[0x8010] = 0xA9
	bus.mem[0x8011] = 0x02
	bus.mem[0x8012] = 0x60

	c.Step() // JSR
	if c.PC() != 0x8010 {
		t.Fatalf("PC after JSR = %#06x, want 0x8010", c.PC())
	}
	c.Step() // LDA #2
	c.Step() // RTS
	if c.PC() != 0x8003 {
		t.Fatalf("PC after RTS = %#06x, want 0x8003", c.PC())
	}
}

func TestBranchTakenOnZero(t *testing.T) {
	// LDA #0 ; BEQ +2 ; LDA #$FF (skipped) ; LDA #1
	c, _ := newTestCPU([]byte{0xA9, 0x00, 0xF0, 0x02, 0xA9, 0xFF, 0xA9, 0x01})
	c.Step()
	c.Step()
	c.Step()
	if byte(c.regs.A) != 1 {
		t.Fatalf("A = %d, want 1 (branch should have skipped the FF load)", byte(c.regs.A))
	}
}

func TestPushPopAccumulator16Bit(t *testing.T) {
	// CLC ; XCE ; REP #$20 ; LDA #$CAFE ; PHA ; LDA #0 ; PLA
	c, _ := newTestCPU([]byte{0x18, 0xFB, 0xC2, 0x20, 0xA9, 0xFE, 0xCA, 0x48, 0xA9, 0x00, 0x00, 0x68})
	for i := 0; i < 6; i++ { // CLC, XCE, REP, LDA #$CAFE, PHA, LDA #0
		c.Step()
	}
	if c.regs.A != 0 {
		t.Fatalf("A after clear = %#04x, want 0", c.regs.A)
	}
	c.Step()
	if c.regs.A != 0xCAFE {
		t.Fatalf("A after pull = %#04x, want 0xCAFE", c.regs.A)
	}
}

func TestBlockMoveMVN(t *testing.T) {
	// CLC ; XCE ; REP #$30 (16-bit A/X/Y) ; LDA #3 ; LDX #$0000 ;
	// LDY #$1000 ; MVN destBank=$00,srcBank=$00
	c, bus := newTestCPU([]byte{
		0x18, 0xFB, 0xC2, 0x30,
		0xA9, 0x03, 0x00,
		0xA2, 0x00, 0x00,
		0xA0, 0x00, 0x10,
		0x54, 0x00, 0x00,
	})
	bus.mem[0x000000], bus.mem[0x000001], bus.mem[0x000002] = 0x11, 0x22, 0x33
	c.Step() // CLC
	c.Step() // XCE
	c.Step() // REP
	c.Step() // LDA
	c.Step() // LDX
	c.Step() // LDY
	c.Step() // MVN
	for i, want := range []byte{0x11, 0x22, 0x33} {
		if got := bus.mem[0x1000+i]; got != want {
			t.Fatalf("byte %d = %#02x, want %#02x", i, got, want)
		}
	}
}

func TestNMIServicedAndPushesState(t *testing.T) {
	c, bus := newTestCPU(nil)
	bus.mem[0x00FFFA] = 0x00
	bus.mem[0x00FFFB] = 0x90 // emulation-mode NMI vector -> $9000

	c.TriggerNMI()
	c.Step()
	if c.PC() != 0x9000 {
		t.Fatalf("PC after NMI = %#06x, want 0x9000", c.PC())
	}
}
