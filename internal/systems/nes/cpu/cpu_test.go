package cpu

import "testing"

// testBus is a flat 64KB address space, enough to exercise the CPU in
// isolation without the rest of the NES memory map.
type testBus struct {
	mem [65536]byte
}

func (b *testBus) Read(addr uint16) byte     { return b.mem[addr] }
func (b *testBus) Write(addr uint16, v byte) { b.mem[addr] = v }

func newTestCPU(resetPC uint16, program []byte) (*CPU, *testBus) {
	bus := &testBus{}
	bus.mem[resetVector] = byte(resetPC)
	bus.mem[resetVector+1] = byte(resetPC >> 8)
	copy(bus.mem[resetPC:], program)

	c := New(bus)
	c.Reset()
	return c, bus
}

func TestResetLoadsPCFromResetVector(t *testing.T) {
	c, _ := newTestCPU(0x8000, nil)
	if c.PC() != 0x8000 {
		t.Fatalf("PC = %#04x, want 0x8000", c.PC())
	}
	if c.regs.SP != 0xFD {
		t.Fatalf("SP = %#02x, want 0xFD", c.regs.SP)
	}
}

func TestLDAImmediateSetsAccumulatorAndFlags(t *testing.T) {
	c, _ := newTestCPU(0x8000, []byte{0xA9, 0x00}) // LDA #$00
	cycles := c.Step()

	if c.regs.A != 0 {
		t.Fatalf("A = %#02x, want 0x00", c.regs.A)
	}
	if !c.regs.getFlag(FlagZero) {
		t.Fatal("expected Zero flag set for LDA #$00")
	}
	if cycles != 2 {
		t.Fatalf("cycles = %d, want 2", cycles)
	}
}

func TestLDAAbsoluteXPageCrossCostsExtraCycle(t *testing.T) {
	c, bus := newTestCPU(0x8000, []byte{0xBD, 0xFF, 0x00}) // LDA $00FF,X
	c.regs.X = 1                                           // crosses into page $0100
	bus.mem[0x0100] = 0x42

	cycles := c.Step()
	if c.regs.A != 0x42 {
		t.Fatalf("A = %#02x, want 0x42", c.regs.A)
	}
	if cycles != 5 {
		t.Fatalf("cycles = %d, want 5 (4 base + 1 page cross)", cycles)
	}
}

func TestADCSetsCarryAndOverflow(t *testing.T) {
	// 0x50 + 0x50 = 0xA0: no unsigned carry, but a signed overflow
	// (two positives producing a negative result).
	c, _ := newTestCPU(0x8000, []byte{0x69, 0x50}) // ADC #$50
	c.regs.A = 0x50

	c.Step()
	if c.regs.A != 0xA0 {
		t.Fatalf("A = %#02x, want 0xA0", c.regs.A)
	}
	if c.regs.getFlag(FlagCarry) {
		t.Fatal("did not expect Carry")
	}
	if !c.regs.getFlag(FlagOverflow) {
		t.Fatal("expected Overflow")
	}
	if !c.regs.getFlag(FlagNegative) {
		t.Fatal("expected Negative")
	}
}

func TestSBCBorrow(t *testing.T) {
	c, _ := newTestCPU(0x8000, []byte{0x38, 0xE9, 0x01}) // SEC ; SBC #$01
	c.regs.A = 0x00

	c.Step() // SEC
	c.Step() // SBC #$01

	if c.regs.A != 0xFF {
		t.Fatalf("A = %#02x, want 0xFF", c.regs.A)
	}
	if c.regs.getFlag(FlagCarry) {
		t.Fatal("expected Carry clear (borrow occurred)")
	}
}

func TestBranchTakenAndNotTaken(t *testing.T) {
	// BEQ +2 ; BRK ; BRK ; target: LDA #$99
	c, _ := newTestCPU(0x8000, []byte{0xF0, 0x02, 0x00, 0x00, 0xA9, 0x99})
	c.regs.setFlag(FlagZero, true)

	cycles := c.Step()
	if cycles != 3 {
		t.Fatalf("taken branch cycles = %d, want 3 (2 base + 1 taken)", cycles)
	}
	if c.PC() != 0x8004 {
		t.Fatalf("PC after taken branch = %#04x, want 0x8004", c.PC())
	}
}

func TestBranchNotTakenStaysSequential(t *testing.T) {
	c, _ := newTestCPU(0x8000, []byte{0xF0, 0x02}) // BEQ +2, Zero clear
	c.regs.setFlag(FlagZero, false)

	cycles := c.Step()
	if cycles != 2 {
		t.Fatalf("not-taken branch cycles = %d, want 2", cycles)
	}
	if c.PC() != 0x8002 {
		t.Fatalf("PC = %#04x, want 0x8002", c.PC())
	}
}

func TestJSRThenRTSRoundTrips(t *testing.T) {
	// JSR $8010 ; (return here) LDA #$AA
	// at $8010: RTS
	c, bus := newTestCPU(0x8000, []byte{0x20, 0x10, 0x80, 0xA9, 0xAA})
	bus.mem[0x8010] = 0x60 // RTS

	c.Step() // JSR
	if c.PC() != 0x8010 {
		t.Fatalf("PC after JSR = %#04x, want 0x8010", c.PC())
	}

	c.Step() // RTS
	if c.PC() != 0x8003 {
		t.Fatalf("PC after RTS = %#04x, want 0x8003 (back after the JSR operand)", c.PC())
	}

	c.Step() // LDA #$AA
	if c.regs.A != 0xAA {
		t.Fatalf("A = %#02x, want 0xAA", c.regs.A)
	}
}

func TestStackPushPopRoundTrips(t *testing.T) {
	c, _ := newTestCPU(0x8000, []byte{0xA9, 0x7E, 0x48, 0xA9, 0x00, 0x68}) // LDA #$7E; PHA; LDA #$00; PLA
	c.Step()
	c.Step()
	c.Step()
	if c.regs.A != 0 {
		t.Fatalf("A after LDA #$00 = %#02x, want 0x00", c.regs.A)
	}
	c.Step() // PLA
	if c.regs.A != 0x7E {
		t.Fatalf("A after PLA = %#02x, want 0x7E", c.regs.A)
	}
}

func TestIllegalOpcodeDoesNotPanic(t *testing.T) {
	c, _ := newTestCPU(0x8000, []byte{0x02}) // unofficial opcode
	c.Step()
}
