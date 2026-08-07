package cpu

import "testing"

type testBus struct {
	mem [0x10000]byte
}

func (b *testBus) Read8(addr uint32) byte { return b.mem[addr&0xFFFF] }
func (b *testBus) Read16(addr uint32) uint16 {
	i := addr & 0xFFFF
	return uint16(b.mem[i]) | uint16(b.mem[i+1])<<8
}
func (b *testBus) Write8(addr uint32, v byte) { b.mem[addr&0xFFFF] = v }
func (b *testBus) Write16(addr uint32, v uint16) {
	i := addr & 0xFFFF
	b.mem[i], b.mem[i+1] = byte(v), byte(v>>8)
}

func newTestCPU(program []byte) (*CPU, *testBus) {
	bus := &testBus{}
	copy(bus.mem[entryPoint:], program)
	return New(bus), bus
}

func TestResetStartsAtEntryPoint(t *testing.T) {
	c, _ := newTestCPU(nil)
	if c.PC() != entryPoint {
		t.Fatalf("PC = %#06x, want %#06x", c.PC(), entryPoint)
	}
}

func TestLoadImmediateAndRegisterToRegister(t *testing.T) {
	// LD A,#$42 ; LD B,A (dst=B(code2)<<4|src=A(code1))
	c, _ := newTestCPU([]byte{0x11, 0x42, 0x48, 0x21})
	c.Step()
	c.Step()
	if c.regs.reg8(1) != 0x42 {
		t.Fatalf("A = %#02x, want 0x42", c.regs.reg8(1))
	}
	if c.regs.reg8(2) != 0x42 {
		t.Fatalf("B = %#02x, want 0x42", c.regs.reg8(2))
	}
}

func TestMemoryLoadStoreViaHL(t *testing.T) {
	// LD HL,#$3000 ; LD A,#$99 ; LD (HL),A ; LD B,(HL)
	c, _ := newTestCPU([]byte{
		0x33, 0x00, 0x30,
		0x11, 0x99,
		0x28 | 1, // LD (HL),A -> reg8 code for A is 1
		0x20 | 2, // LD B,(HL) -> reg8 code for B is 2
	})
	c.Step()
	c.Step()
	c.Step()
	c.Step()
	if c.regs.reg8(2) != 0x99 {
		t.Fatalf("B = %#02x, want 0x99", c.regs.reg8(2))
	}
	if v := c.bus.Read8(0x3000); v != 0x99 {
		t.Fatalf("mem[0x3000] = %#02x, want 0x99", v)
	}
}

func TestALUAddSetsFlags(t *testing.T) {
	// LD A,#$F0 ; ADD A,#$20 -> 0x110 truncates to 0x10, Carry set
	c, _ := newTestCPU([]byte{0x11, 0xF0, 0x50, 0x20})
	c.Step()
	c.Step()
	if c.regs.reg8(1) != 0x10 {
		t.Fatalf("A = %#02x, want 0x10", c.regs.reg8(1))
	}
	if !c.regs.getFlag(FlagCarry) {
		t.Fatal("expected Carry set")
	}
}

func TestJumpConditionalOnZero(t *testing.T) {
	// LD A,#0 ; CP A,#0 (op index 7 within 0x50 range) ; JP Z,#imm ; LD A,#$FF (skipped)
	c, _ := newTestCPU([]byte{
		0x11, 0x00,
		0x57, 0x00, // CP A,#0 -> Z set
		0x91, 0x09, 0x00, // JP Z,$0009
		0x11, 0xFF,
		0x11, 0x01, // at $0009
	})
	c.Step() // LD A,#0
	c.Step() // CP
	c.Step() // JP Z
	if c.PC() != 0x0009 {
		t.Fatalf("PC = %#06x, want 0x0009", c.PC())
	}
	c.Step()
	if c.regs.reg8(1) != 1 {
		t.Fatalf("A = %d, want 1", c.regs.reg8(1))
	}
}

func TestCallAndReturn(t *testing.T) {
	// LD SP,#$4000 ; CALL $0010 ; LD A,#$01 ; at $0010: LD A,#$02 ; RET
	c, bus := newTestCPU([]byte{
		0x3B, 0x00, 0x40,
		0xB0, 0x10, 0x00,
		0x11, 0x01,
	})
	bus.mem[0x10] = 0x11
	bus.mem[0x11] = 0x02
	bus.mem[0x12] = 0xB8

	c.Step() // LD SP
	c.Step() // CALL
	if c.PC() != 0x0010 {
		t.Fatalf("PC after CALL = %#06x, want 0x0010", c.PC())
	}
	c.Step() // LD A,#2
	c.Step() // RET
	if c.PC() != 0x0006 {
		t.Fatalf("PC after RET = %#06x, want 0x0006", c.PC())
	}
}

func TestPushPopRoundTrip(t *testing.T) {
	// LD SP,#$4000 ; LD BC,#$CAFE ; PUSH BC ; LD BC,#0 ; POP BC
	c, _ := newTestCPU([]byte{
		0x3B, 0x00, 0x40,
		0x31, 0xFE, 0xCA,
		0xC0 | 1,
		0x31, 0x00, 0x00,
		0xC4 | 1,
	})
	c.Step()
	c.Step()
	c.Step()
	c.Step()
	if c.regs.reg16(1) != 0 {
		t.Fatalf("BC after clear = %#04x, want 0", c.regs.reg16(1))
	}
	c.Step()
	if c.regs.reg16(1) != 0xCAFE {
		t.Fatalf("BC after pop = %#04x, want 0xCAFE", c.regs.reg16(1))
	}
}

func TestIRQServicedAndReturnsViaRETI(t *testing.T) {
	c, bus := newTestCPU(nil)
	c.regs.XSP = 0x4000
	c.regs.SR &^= 0x7000 // unmask interrupts

	bus.mem[0x0100] = 0xBF // RETI at the vector

	c.TriggerIRQ(0x0100)
	c.Step() // services IRQ, jumps to vector
	if c.PC() != 0x0100 {
		t.Fatalf("PC after IRQ = %#06x, want 0x0100", c.PC())
	}
	c.Step() // RETI
	if c.PC() != entryPoint {
		t.Fatalf("PC after RETI = %#06x, want %#06x", c.PC(), entryPoint)
	}
}
