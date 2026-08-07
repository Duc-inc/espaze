package cpu

import "testing"

type testBus struct {
	mem [65536]byte
}

func (b *testBus) Read(addr uint16) byte     { return b.mem[addr] }
func (b *testBus) Write(addr uint16, v byte) { b.mem[addr] = v }

type testIO struct {
	in  [256]byte
	out [256]byte
}

func (io *testIO) In(port byte) byte     { return io.in[port] }
func (io *testIO) Out(port byte, v byte) { io.out[port] = v }

func newTestCPU(program []byte) (*CPU, *testBus) {
	bus := &testBus{}
	copy(bus.mem[0:], program)
	c := New(bus, &testIO{})
	c.regs.PC = 0
	return c, bus
}

func TestResetState(t *testing.T) {
	c, _ := newTestCPU(nil)
	if c.regs.SP != 0xFFFF {
		t.Fatalf("SP = %#04x, want 0xFFFF", c.regs.SP)
	}
	if c.regs.IFF1 || c.regs.IFF2 {
		t.Fatal("interrupts should start disabled")
	}
}

func TestLD8BitImmediateAndRegisterToRegister(t *testing.T) {
	c, _ := newTestCPU([]byte{0x06, 0x42, 0x78}) // LD B,0x42 ; LD A,B
	c.Step()
	if c.regs.B != 0x42 {
		t.Fatalf("B = %#02x, want 0x42", c.regs.B)
	}
	c.Step()
	if c.regs.A != 0x42 {
		t.Fatalf("A = %#02x, want 0x42", c.regs.A)
	}
}

func TestADDSetsCarryAndOverflow(t *testing.T) {
	c, _ := newTestCPU([]byte{0x3E, 0x50, 0xC6, 0x50}) // LD A,0x50 ; ADD A,0x50
	c.Step()
	c.Step()
	if c.regs.A != 0xA0 {
		t.Fatalf("A = %#02x, want 0xA0", c.regs.A)
	}
	if c.regs.getFlag(FlagC) {
		t.Fatal("did not expect Carry")
	}
	if !c.regs.getFlag(FlagPV) {
		t.Fatal("expected overflow (two positives producing a negative)")
	}
	if !c.regs.getFlag(FlagS) {
		t.Fatal("expected Sign set")
	}
}

func TestCPDoesNotModifyA(t *testing.T) {
	c, _ := newTestCPU([]byte{0x3E, 0x10, 0xFE, 0x10}) // LD A,0x10 ; CP 0x10
	c.Step()
	c.Step()
	if c.regs.A != 0x10 {
		t.Fatalf("A = %#02x, want unchanged 0x10", c.regs.A)
	}
	if !c.regs.getFlag(FlagZ) {
		t.Fatal("expected Zero (equal comparison)")
	}
}

func TestANDSetsParityAndHalfCarry(t *testing.T) {
	c, _ := newTestCPU([]byte{0x3E, 0xFF, 0xE6, 0x0F}) // LD A,0xFF ; AND 0x0F
	c.Step()
	c.Step()
	if c.regs.A != 0x0F {
		t.Fatalf("A = %#02x, want 0x0F", c.regs.A)
	}
	if !c.regs.getFlag(FlagH) {
		t.Fatal("AND always sets Half-carry")
	}
	if !c.regs.getFlag(FlagPV) {
		t.Fatal("expected even parity (0x0F = 4 bits set)")
	}
}

func TestINCDoesNotTouchCarry(t *testing.T) {
	c, _ := newTestCPU([]byte{0x37, 0x3E, 0xFF, 0x3C}) // SCF ; LD A,0xFF ; INC A
	c.Step()
	c.Step()
	c.Step()
	if c.regs.A != 0x00 {
		t.Fatalf("A = %#02x, want 0x00 (wrapped)", c.regs.A)
	}
	if !c.regs.getFlag(FlagC) {
		t.Fatal("INC must not clear a Carry set beforehand")
	}
	if !c.regs.getFlag(FlagZ) {
		t.Fatal("expected Zero")
	}
}

func Test16BitLoadAndAddHL(t *testing.T) {
	c, _ := newTestCPU([]byte{0x21, 0x34, 0x12, 0x01, 0x01, 0x00, 0x09}) // LD HL,0x1234 ; LD BC,0x0001 ; ADD HL,BC
	c.Step()
	c.Step()
	c.Step()
	if c.regs.HL() != 0x1235 {
		t.Fatalf("HL = %#04x, want 0x1235", c.regs.HL())
	}
}

func TestJRRelativeJump(t *testing.T) {
	c, _ := newTestCPU([]byte{0x18, 0x02, 0x00, 0x00, 0x3E, 0x99}) // JR +2 ; NOP;NOP ; LD A,0x99
	c.Step()
	if c.regs.PC != 4 {
		t.Fatalf("PC after JR = %#04x, want 4", c.regs.PC)
	}
	c.Step()
	if c.regs.A != 0x99 {
		t.Fatalf("A = %#02x, want 0x99", c.regs.A)
	}
}

func TestCallAndRetRoundTrip(t *testing.T) {
	// CALL 0x0006 ; (return here) LD A,0xAA ; at 0x0006: RET
	c, bus := newTestCPU([]byte{0xCD, 0x06, 0x00, 0x3E, 0xAA})
	bus.mem[0x0006] = 0xC9 // RET

	c.Step() // CALL
	if c.regs.PC != 6 {
		t.Fatalf("PC after CALL = %#04x, want 6", c.regs.PC)
	}
	c.Step() // RET
	if c.regs.PC != 3 {
		t.Fatalf("PC after RET = %#04x, want 3", c.regs.PC)
	}
	c.Step() // LD A,0xAA
	if c.regs.A != 0xAA {
		t.Fatalf("A = %#02x, want 0xAA", c.regs.A)
	}
}

func TestPushPopRoundTrip(t *testing.T) {
	c, _ := newTestCPU([]byte{0x01, 0x34, 0x12, 0xC5, 0x21, 0x00, 0x00, 0xE1}) // LD BC,0x1234 ; PUSH BC ; LD HL,0 ; POP HL
	c.Step()
	c.Step()
	c.Step()
	c.Step()
	if c.regs.HL() != 0x1234 {
		t.Fatalf("HL after POP = %#04x, want 0x1234", c.regs.HL())
	}
}

func TestEXXAndExAFSwapShadowRegisters(t *testing.T) {
	c, _ := newTestCPU([]byte{0x3E, 0x11, 0x08, 0x06, 0x22, 0x3E, 0x33, 0x08}) // LD A,0x11 ; EX AF,AF' ; LD B,0x22 ; LD A,0x33 ; EX AF,AF'
	c.Step()                                                                   // LD A,0x11
	c.Step()                                                                   // EX AF,AF' -> A2=0x11, A=0(old A2)
	c.Step()                                                                   // LD B,0x22 (irrelevant, just advancing)
	c.Step()                                                                   // LD A,0x33
	c.Step()                                                                   // EX AF,AF' -> A should become 0x11 again
	if c.regs.A != 0x11 {
		t.Fatalf("A after second EX AF,AF' = %#02x, want 0x11", c.regs.A)
	}
}
