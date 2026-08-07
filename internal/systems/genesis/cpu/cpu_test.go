package cpu

import "testing"

type testBus struct {
	mem [0x10000]byte
}

func (b *testBus) Read8(addr uint32) byte { return b.mem[addr&0xFFFF] }
func (b *testBus) Read16(addr uint32) uint16 {
	return uint16(b.mem[addr&0xFFFF])<<8 | uint16(b.mem[(addr+1)&0xFFFF])
}
func (b *testBus) Read32(addr uint32) uint32 {
	return uint32(b.Read16(addr))<<16 | uint32(b.Read16(addr+2))
}
func (b *testBus) Write8(addr uint32, v byte) { b.mem[addr&0xFFFF] = v }
func (b *testBus) Write16(addr uint32, v uint16) {
	b.mem[addr&0xFFFF] = byte(v >> 8)
	b.mem[(addr+1)&0xFFFF] = byte(v)
}
func (b *testBus) Write32(addr uint32, v uint32) {
	b.Write16(addr, uint16(v>>16))
	b.Write16(addr+2, uint16(v))
}

// newTestCPU sets up a reset vector pointing at $1000 (SSP = $8000) and
// writes program starting there.
func newTestCPU(program []byte) (*CPU, *testBus) {
	bus := &testBus{}
	bus.Write32(0, 0x8000) // initial SSP
	bus.Write32(4, 0x1000) // initial PC
	copy(bus.mem[0x1000:], program)
	return New(bus), bus
}

func TestResetLoadsSSPAndPC(t *testing.T) {
	c, _ := newTestCPU(nil)
	if c.PC() != 0x1000 {
		t.Fatalf("PC = %#06x, want 0x1000", c.PC())
	}
	if c.regs.A[7] != 0x8000 {
		t.Fatalf("A7 (SSP) = %#06x, want 0x8000", c.regs.A[7])
	}
	if !c.regs.supervisor() {
		t.Fatal("expected supervisor mode after reset")
	}
}

func TestMoveqLoadsSignExtendedImmediate(t *testing.T) {
	c, _ := newTestCPU([]byte{0x7A, 0xFF}) // MOVEQ #-1,D5
	c.Step()
	if c.regs.D[5] != 0xFFFFFFFF {
		t.Fatalf("D5 = %#010x, want 0xFFFFFFFF", c.regs.D[5])
	}
	if !c.regs.getFlag(FlagN) {
		t.Fatal("expected Negative flag")
	}
}

func TestMoveByteToDataRegisterOnlyTouchesLowByte(t *testing.T) {
	// MOVE.B #$FF,D0 with D0 pre-loaded to 0x12345678 - only the low
	// byte should change.
	c, _ := newTestCPU([]byte{0x10, 0x3C, 0x00, 0xFF}) // MOVE.B #$FF,D0
	c.regs.D[0] = 0x12345678
	c.Step()
	if c.regs.D[0] != 0x123456FF {
		t.Fatalf("D0 = %#010x, want 0x123456ff (only low byte replaced)", c.regs.D[0])
	}
}

func TestMoveToMemoryAndBack(t *testing.T) {
	// MOVE.L D0,$2000.L ; MOVE.L $2000.L,D1 (mode 7/reg 1 = absolute
	// long - reg 0 there would mean absolute *short*, a 16-bit address).
	c, _ := newTestCPU([]byte{0x23, 0xC0, 0x00, 0x00, 0x20, 0x00, 0x22, 0x39, 0x00, 0x00, 0x20, 0x00})
	c.regs.D[0] = 0xCAFEBABE
	c.Step()
	c.Step()
	if c.regs.D[1] != 0xCAFEBABE {
		t.Fatalf("D1 = %#010x, want 0xCAFEBABE", c.regs.D[1])
	}
}

func TestLEALoadsEffectiveAddress(t *testing.T) {
	c, _ := newTestCPU([]byte{0x41, 0xF9, 0x00, 0x00, 0x30, 0x00}) // LEA $3000,A0
	c.Step()
	if c.regs.A[0] != 0x3000 {
		t.Fatalf("A0 = %#06x, want 0x3000", c.regs.A[0])
	}
}

func TestAddSetsCarryAndOverflow(t *testing.T) {
	// MOVE.B #$50,D0 ; ADD.B #$50,D0
	c, _ := newTestCPU([]byte{0x10, 0x3C, 0x00, 0x50, 0x06, 0x00, 0x00, 0x50})
	c.Step()
	c.Step()
	if byte(c.regs.D[0]) != 0xA0 {
		t.Fatalf("D0 low byte = %#02x, want 0xA0", byte(c.regs.D[0]))
	}
	if c.regs.getFlag(FlagC) {
		t.Fatal("did not expect Carry")
	}
	if !c.regs.getFlag(FlagV) {
		t.Fatal("expected overflow (two positives producing a negative)")
	}
}

func TestSubqTargetingAddressRegisterSkipsFlags(t *testing.T) {
	c, _ := newTestCPU([]byte{0x53, 0x4F}) // SUBQ.W #1,A7
	c.regs.SR |= FlagZ                     // pre-set Z to prove SUBQ-to-An leaves it alone
	before := c.regs.A[7]
	c.Step()
	if c.regs.A[7] != before-1 {
		t.Fatalf("A7 = %#06x, want %#06x", c.regs.A[7], before-1)
	}
	if !c.regs.getFlag(FlagZ) {
		t.Fatal("SUBQ targeting an address register must not touch flags")
	}
}

func TestCmpSetsZeroWithoutModifyingOperand(t *testing.T) {
	// MOVEQ #5,D0 ; CMPI.L #5,D0
	c, _ := newTestCPU([]byte{0x70, 0x05, 0x0C, 0x80, 0x00, 0x00, 0x00, 0x05})
	c.Step()
	c.Step()
	if c.regs.D[0] != 5 {
		t.Fatalf("D0 = %d, want unchanged 5", c.regs.D[0])
	}
	if !c.regs.getFlag(FlagZ) {
		t.Fatal("expected Zero flag from an equal comparison")
	}
}

func TestAndOrEor(t *testing.T) {
	// MOVEQ #$0F,D0 ; AND.B #$03,D0 ; OR.B #$F0,D0 ; EOR.B #$FF,D0
	c, _ := newTestCPU([]byte{
		0x70, 0x0F,
		0x02, 0x00, 0x00, 0x03,
		0x00, 0x00, 0x00, 0xF0,
		0x0A, 0x00, 0x00, 0xFF,
	})
	c.Step() // MOVEQ
	c.Step() // AND -> 0x03
	if byte(c.regs.D[0]) != 0x03 {
		t.Fatalf("after AND, D0 low byte = %#02x, want 0x03", byte(c.regs.D[0]))
	}
	c.Step() // OR -> 0xF3
	if byte(c.regs.D[0]) != 0xF3 {
		t.Fatalf("after OR, D0 low byte = %#02x, want 0xF3", byte(c.regs.D[0]))
	}
	c.Step() // EOR -> 0x0C
	if byte(c.regs.D[0]) != 0x0C {
		t.Fatalf("after EOR, D0 low byte = %#02x, want 0x0C", byte(c.regs.D[0]))
	}
}

func TestShiftLeftSetsCarryFromMSB(t *testing.T) {
	// MOVEQ #-1,D0 (all bits set) ; LSL.B #1,D0
	c, _ := newTestCPU([]byte{0x70, 0xFF, 0xE3, 0x08})
	c.Step()
	c.Step()
	if byte(c.regs.D[0]) != 0xFE {
		t.Fatalf("D0 low byte = %#02x, want 0xFE", byte(c.regs.D[0]))
	}
	if !c.regs.getFlag(FlagC) {
		t.Fatal("expected Carry set from the bit shifted out")
	}
}

func TestBtstBsetOnDataRegister(t *testing.T) {
	// MOVEQ #0,D0 ; BSET #3,D0 ; BTST #3,D0
	c, _ := newTestCPU([]byte{0x70, 0x00, 0x08, 0xC0, 0x00, 0x03, 0x08, 0x00, 0x00, 0x03})
	c.Step() // MOVEQ
	c.Step() // BSET #3,D0
	if c.regs.D[0] != 0x08 {
		t.Fatalf("D0 = %#02x, want 0x08", c.regs.D[0])
	}
	c.Step() // BTST #3,D0
	if c.regs.getFlag(FlagZ) {
		t.Fatal("BTST should find the bit set (Z clear)")
	}
}
