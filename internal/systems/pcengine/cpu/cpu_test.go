package cpu

import "testing"

type testBus struct {
	mem [0x200000]byte
}

func (b *testBus) Read(addr uint32) byte     { return b.mem[addr&0x1FFFFF] }
func (b *testBus) Write(addr uint32, v byte) { b.mem[addr&0x1FFFFF] = v }

// newTestCPU relies on every MPR page defaulting to 0 (see CPU.Reset),
// so logical $0000-$1FFF and the reset vector at logical $FFFE (physical
// $1FFE) both land in the same physical page 0 the program is written to.
func newTestCPU(program []byte) (*CPU, *testBus) {
	bus := &testBus{}
	bus.mem[0x1FFE] = 0x00 // reset vector low byte
	bus.mem[0x1FFF] = 0x00 // reset vector high byte -> logical/physical $0000
	copy(bus.mem[0:], program)
	return New(bus), bus
}

func TestResetLoadsPCFromVector(t *testing.T) {
	c, _ := newTestCPU(nil)
	if c.PC() != 0x0000 {
		t.Fatalf("PC = %#04x, want 0x0000", c.PC())
	}
}

func TestLDAImmediateAndSTAAbsoluteRoundTrip(t *testing.T) {
	// LDA #$42 ; STA $2000 ; LDA #$00 ; LDA $2000
	c, _ := newTestCPU([]byte{0xA9, 0x42, 0x8D, 0x00, 0x20, 0xA9, 0x00, 0xAD, 0x00, 0x20})
	c.Step()
	c.Step()
	c.Step()
	c.Step()
	if c.regs.A != 0x42 {
		t.Fatalf("A = %#02x, want 0x42", c.regs.A)
	}
}

func TestAdcIsAlwaysBinaryRegardlessOfDecimalFlag(t *testing.T) {
	// SED ; LDA #$09 ; ADC #$01 (would be 0x10 in BCD, 0x0A in binary)
	c, _ := newTestCPU([]byte{0xF8, 0xA9, 0x09, 0x69, 0x01})
	c.Step()
	c.Step()
	c.Step()
	if c.regs.A != 0x0A {
		t.Fatalf("A = %#02x, want 0x0A (HuC6280 ADC ignores Decimal)", c.regs.A)
	}
}

func TestBranchTakenAndNotTaken(t *testing.T) {
	// LDA #$00 ; BEQ +2 ; LDA #$FF (skipped) ; LDA #$01
	c, _ := newTestCPU([]byte{0xA9, 0x00, 0xF0, 0x02, 0xA9, 0xFF, 0xA9, 0x01})
	c.Step()
	c.Step()
	c.Step()
	if c.regs.A != 0x01 {
		t.Fatalf("A = %#02x, want 0x01 (branch should have skipped the FF load)", c.regs.A)
	}
}

func TestJsrRts(t *testing.T) {
	// JSR $0010 ; (return here) LDA #$01 ; at $0010: LDA #$02 ; RTS
	c, bus := newTestCPU([]byte{0x20, 0x10, 0x00, 0xA9, 0x01})
	bus.mem[0x10], bus.mem[0x11] = 0xA9, 0x02
	bus.mem[0x12] = 0x60

	c.Step() // JSR
	if c.PC() != 0x0010 {
		t.Fatalf("PC after JSR = %#04x, want 0x0010", c.PC())
	}
	c.Step() // LDA #$02
	c.Step() // RTS
	if c.PC() != 0x0003 {
		t.Fatalf("PC after RTS = %#04x, want 0x0003", c.PC())
	}
}

func TestTamRemapsLogicalPage(t *testing.T) {
	// LDA #$05 ; TAM #$01 (maps MPR0, logical $0000-$1FFF, to physical page 5)
	c, bus := newTestCPU([]byte{0xA9, 0x05, 0x53, 0x01})
	bus.mem[5*0x2000+0x0100] = 0x77 // a byte in physical page 5

	c.Step() // LDA #$05
	c.Step() // TAM #$01
	if c.mmu.mpr[0] != 0x05 {
		t.Fatalf("MPR0 = %#02x, want 0x05", c.mmu.mpr[0])
	}
	if v := c.read(0x0100); v != 0x77 {
		t.Fatalf("read through remapped MPR0 = %#02x, want 0x77", v)
	}
}

func TestTiiBlockTransfer(t *testing.T) {
	// TII $2000,$4000,$0004 (copy 4 bytes); MPR1/MPR2 are mapped to
	// distinct physical pages first so source/dest don't alias the
	// default all-zero mapping (or each other, or the program at page 0).
	c, bus := newTestCPU([]byte{0x73, 0x00, 0x20, 0x00, 0x40, 0x04, 0x00})
	c.mmu.mpr[1] = 1 // logical $2000-$3FFF -> physical page 1
	c.mmu.mpr[2] = 2 // logical $4000-$5FFF -> physical page 2
	copy(bus.mem[0x2000:], []byte{0x11, 0x22, 0x33, 0x44})

	c.Step()
	got := bus.mem[0x4000 : 0x4000+4]
	want := []byte{0x11, 0x22, 0x33, 0x44}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("byte %d = %#02x, want %#02x", i, got[i], want[i])
		}
	}
}

func TestIRQ2ServicedThroughVector(t *testing.T) {
	c, bus := newTestCPU(nil)
	bus.mem[0x1FF6] = 0x00 // IRQ2 vector low
	bus.mem[0x1FF7] = 0x40 // IRQ2 vector high -> $4000
	c.regs.setFlag(FlagInterrupt, false)

	c.TriggerIRQ2()
	c.Step()
	if c.PC() != 0x4000 {
		t.Fatalf("PC after IRQ2 = %#04x, want 0x4000", c.PC())
	}
}

func TestIRQMaskBlocksService(t *testing.T) {
	c, _ := newTestCPU([]byte{0xEA}) // NOP
	c.regs.setFlag(FlagInterrupt, false)
	c.WriteIRQMask(0x01) // mask IRQ2

	c.TriggerIRQ2()
	c.Step()
	if c.PC() != 0x0001 {
		t.Fatalf("PC = %#04x, want 0x0001 (masked IRQ2 should not have been serviced)", c.PC())
	}
}
