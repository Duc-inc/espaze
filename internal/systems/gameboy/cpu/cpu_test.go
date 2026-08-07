package cpu

import "testing"

// flatBus is a trivial 64KB RAM used only by these tests - no MMU
// routing, no I/O side effects, so instruction semantics can be checked
// in isolation from the rest of the hardware.
type flatBus struct {
	mem [0x10000]byte
}

func (b *flatBus) Read(addr uint16) byte     { return b.mem[addr] }
func (b *flatBus) Write(addr uint16, v byte) { b.mem[addr] = v }

func newTestCPU(program ...byte) (*CPU, *flatBus) {
	bus := &flatBus{}
	copy(bus.mem[0x0100:], program)
	c := New(bus)
	return c, bus
}

func TestOpcodeTablesFullyPopulated(t *testing.T) {
	for i := 0; i < 256; i++ {
		if i != 0xCB && mainTable[i] == nil { // 0xCB is the CB-prefix escape, never dispatched directly
			t.Errorf("mainTable[0x%02X] is nil", i)
		}
		if cbTable[i] == nil {
			t.Errorf("cbTable[0x%02X] is nil", i)
		}
	}
}

func TestLoadImmediateAndRegisterToRegister(t *testing.T) {
	c, _ := newTestCPU(0x06, 0x42, 0x41) // LD B,0x42 ; LD B,C
	c.regs.C = 0x00
	c.Step()
	if c.regs.B != 0x42 {
		t.Fatalf("LD B,0x42: got B=0x%02X", c.regs.B)
	}
	c.Step()
	if c.regs.B != 0x00 {
		t.Fatalf("LD B,C: got B=0x%02X, want 0x00", c.regs.B)
	}
}

func TestAddSetsFlagsCorrectly(t *testing.T) {
	c, _ := newTestCPU(0x3E, 0xFF, 0xC6, 0x01) // LD A,0xFF ; ADD A,0x01
	c.Step()
	c.Step()
	if c.regs.A != 0x00 {
		t.Fatalf("ADD A,0x01: got A=0x%02X, want 0x00", c.regs.A)
	}
	if !c.regs.HasFlag(FlagZ) || !c.regs.HasFlag(FlagH) || !c.regs.HasFlag(FlagC) {
		t.Fatalf("ADD overflow flags: F=0x%02X, want Z+H+C set", c.regs.F)
	}
}

func TestIncDecFlagsLeaveCarryAlone(t *testing.T) {
	c, _ := newTestCPU(0x37, 0x3C, 0x3D) // SCF ; INC A ; DEC A
	c.regs.A = 0x00
	c.Step()
	if !c.regs.HasFlag(FlagC) {
		t.Fatalf("SCF should set carry")
	}
	c.Step() // INC A: 0 -> 1
	if !c.regs.HasFlag(FlagC) {
		t.Fatalf("INC must not touch the carry flag")
	}
	c.Step() // DEC A: 1 -> 0
	if !c.regs.HasFlag(FlagZ) {
		t.Fatalf("DEC A back to 0 should set Z")
	}
	if !c.regs.HasFlag(FlagC) {
		t.Fatalf("DEC must not touch the carry flag")
	}
}

func TestPushPopRoundTrip(t *testing.T) {
	c, _ := newTestCPU(0x01, 0x34, 0x12, 0xC5, 0xD1) // LD BC,0x1234 ; PUSH BC ; POP DE
	c.regs.SP = 0xFFFE
	c.Step()
	c.Step()
	c.Step()
	if c.regs.DE() != 0x1234 {
		t.Fatalf("PUSH BC/POP DE: got DE=0x%04X, want 0x1234", c.regs.DE())
	}
}

func TestCallAndReturn(t *testing.T) {
	c, bus := newTestCPU(0xCD, 0x00, 0x02) // CALL 0x0200
	bus.mem[0x0200] = 0xC9                 // RET
	c.regs.SP = 0xFFFE
	c.Step() // CALL
	if c.regs.PC != 0x0200 {
		t.Fatalf("CALL: PC=0x%04X, want 0x0200", c.regs.PC)
	}
	c.Step() // RET
	if c.regs.PC != 0x0103 {
		t.Fatalf("RET: PC=0x%04X, want 0x0103 (right after the 3-byte CALL)", c.regs.PC)
	}
}

func TestConditionalJumpRelative(t *testing.T) {
	c, _ := newTestCPU(0xAF, 0x20, 0x02, 0x00, 0x00, 0x3E, 0x99) // XOR A ; JR NZ,+2 ; NOP;NOP ; LD A,0x99
	c.Step()                                                     // XOR A -> A=0, Z set
	c.Step()                                                     // JR NZ: Z is set, so NOT taken
	if c.regs.PC != 0x0103 {
		t.Fatalf("JR NZ should not have jumped: PC=0x%04X", c.regs.PC)
	}
}

func TestCbBitResSet(t *testing.T) {
	c, _ := newTestCPU(0x3E, 0x00, 0xCB, 0xC7, 0xCB, 0x47, 0xCB, 0x87)
	// LD A,0x00 ; SET 0,A ; BIT 0,A ; RES 0,A
	c.Step() // LD A,0
	c.Step() // SET 0,A -> A=0x01
	if c.regs.A != 0x01 {
		t.Fatalf("SET 0,A: got A=0x%02X, want 0x01", c.regs.A)
	}
	c.Step() // BIT 0,A -> Z should be false (bit is set)
	if c.regs.HasFlag(FlagZ) {
		t.Fatalf("BIT 0,A: Z should be clear, bit 0 is set")
	}
	c.Step() // RES 0,A -> A=0x00
	if c.regs.A != 0x00 {
		t.Fatalf("RES 0,A: got A=0x%02X, want 0x00", c.regs.A)
	}
}

func TestInterruptDispatchAndReturn(t *testing.T) {
	c, bus := newTestCPU(0xFB, 0x00) // EI ; NOP
	c.regs.SP = 0xFFFE
	bus.mem[0x0040] = 0xD9  // RETI at the VBlank vector
	bus.Write(0xFFFF, 0x01) // IE: VBlank enabled
	bus.Write(0xFF0F, 0x01) // IF: VBlank requested

	c.Step() // EI (delay=2, ime still false)
	c.Step() // NOP (delay=1->0, ime becomes true at top of *next* step)
	c.Step() // interrupt should dispatch now
	if c.regs.PC != 0x0040 {
		t.Fatalf("expected interrupt dispatch to 0x0040, got PC=0x%04X", c.regs.PC)
	}
	if bus.Read(0xFF0F)&0x01 != 0 {
		t.Fatalf("IF bit should be cleared once the interrupt is dispatched")
	}
	c.Step() // RETI
	if c.regs.PC != 0x0102 || !c.ime {
		t.Fatalf("RETI should return to 0x0102 with IME re-enabled: PC=0x%04X ime=%v", c.regs.PC, c.ime)
	}
}
