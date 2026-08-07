package spc700

import "testing"

type testBus struct {
	mem [0x10000]byte
}

func (b *testBus) Read8(addr uint16) byte     { return b.mem[addr] }
func (b *testBus) Write8(addr uint16, v byte) { b.mem[addr] = v }

func newTestCPU(program []byte) (*CPU, *testBus) {
	bus := &testBus{}
	copy(bus.mem[entryPoint:], program)
	return New(bus), bus
}

func TestResetStartsAtEntryPoint(t *testing.T) {
	c, _ := newTestCPU(nil)
	if c.PC() != entryPoint {
		t.Fatalf("PC = %#04x, want %#04x", c.PC(), entryPoint)
	}
}

func TestLoadImmediateAndMove(t *testing.T) {
	// LD A,#$42 ; MOV X,A (dst=X(1)<<4|src=A(0))
	c, _ := newTestCPU([]byte{0x10, 0x42, 0x28, 0x10})
	c.Step()
	c.Step()
	if c.regs.X != 0x42 {
		t.Fatalf("X = %#02x, want 0x42", c.regs.X)
	}
}

func TestDirectPageLoadStore(t *testing.T) {
	// LD A,#$99 ; LD (dp=$10),A ; LD A,#0 ; LD A,(dp=$10)
	c, _ := newTestCPU([]byte{0x10, 0x99, 0x1C, 0x10, 0x10, 0x00, 0x18, 0x10})
	c.Step()
	c.Step()
	c.Step()
	c.Step()
	if c.regs.A != 0x99 {
		t.Fatalf("A = %#02x, want 0x99", c.regs.A)
	}
}

func TestALUAdd(t *testing.T) {
	// LD A,#$F0 ; ADD A,#$20 -> wraps to 0x10, Carry set
	c, _ := newTestCPU([]byte{0x10, 0xF0, 0x30, 0x20})
	c.Step()
	c.Step()
	if c.regs.A != 0x10 {
		t.Fatalf("A = %#02x, want 0x10", c.regs.A)
	}
	if !c.regs.getFlag(FlagCarry) {
		t.Fatal("expected Carry set")
	}
}

func TestBranchOnEqual(t *testing.T) {
	// LD A,#0 ; CMP A,#0 (op index 4) ; BEQ +2 ; LD A,#$FF (skipped) ; LD A,#1
	c, _ := newTestCPU([]byte{0x10, 0x00, 0x34, 0x00, 0x50, 0x02, 0x10, 0xFF, 0x10, 0x01})
	c.Step() // LD A,#0
	c.Step() // CMP
	c.Step() // BEQ
	if c.PC() != entryPoint+8 {
		t.Fatalf("PC = %#04x, want %#04x", c.PC(), entryPoint+8)
	}
}

func TestCallAndReturn(t *testing.T) {
	// CALL $FFD0 ; LD A,#1 ; at $FFD0: LD A,#2 ; RET
	c, bus := newTestCPU([]byte{0x60, 0xD0, 0xFF, 0x10, 0x01})
	bus.mem[0xFFD0] = 0x10
	bus.mem[0xFFD1] = 0x02
	bus.mem[0xFFD2] = 0x61

	c.Step() // CALL
	if c.PC() != 0xFFD0 {
		t.Fatalf("PC after CALL = %#04x, want 0xFFD0", c.PC())
	}
	c.Step() // LD A,#2
	c.Step() // RET
	if c.PC() != entryPoint+3 {
		t.Fatalf("PC after RET = %#04x, want %#04x", c.PC(), entryPoint+3)
	}
}

func TestPushPopRoundTrip(t *testing.T) {
	// LD A,#$77 ; PUSH A ; LD A,#0 ; POP A
	c, _ := newTestCPU([]byte{0x10, 0x77, 0x68, 0x10, 0x00, 0x6C})
	c.Step()
	c.Step()
	c.Step()
	if c.regs.A != 0 {
		t.Fatalf("A after clear = %#02x, want 0", c.regs.A)
	}
	c.Step()
	if c.regs.A != 0x77 {
		t.Fatalf("A after pop = %#02x, want 0x77", c.regs.A)
	}
}
