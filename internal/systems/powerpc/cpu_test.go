package powerpc

import "testing"

type testBus struct {
	mem [0x10000]byte
}

func (b *testBus) Read8(addr uint32) byte { return b.mem[addr&0xFFFF] }
func (b *testBus) Read16(addr uint32) uint16 {
	i := addr & 0xFFFF
	return uint16(b.mem[i])<<8 | uint16(b.mem[i+1])
}
func (b *testBus) Read32(addr uint32) uint32 {
	i := addr & 0xFFFF
	return uint32(b.mem[i])<<24 | uint32(b.mem[i+1])<<16 | uint32(b.mem[i+2])<<8 | uint32(b.mem[i+3])
}
func (b *testBus) Write8(addr uint32, v byte) { b.mem[addr&0xFFFF] = v }
func (b *testBus) Write16(addr uint32, v uint16) {
	i := addr & 0xFFFF
	b.mem[i], b.mem[i+1] = byte(v>>8), byte(v)
}
func (b *testBus) Write32(addr uint32, v uint32) {
	i := addr & 0xFFFF
	b.mem[i], b.mem[i+1], b.mem[i+2], b.mem[i+3] = byte(v>>24), byte(v>>16), byte(v>>8), byte(v)
}

func newTestCPU(program []uint32) (*CPU, *testBus) {
	bus := &testBus{}
	for i, instr := range program {
		bus.Write32(uint32(i*4), instr)
	}
	return New(bus), bus
}

func TestResetStartsAtZero(t *testing.T) {
	c, _ := newTestCPU(nil)
	if c.PC() != 0 {
		t.Fatalf("PC = %#08x, want 0", c.PC())
	}
}

func TestAddiLoadsImmediate(t *testing.T) {
	// addi r3, r0, 42  -> opcode14, rD=3, rA=0, SIMM=42
	instr := uint32(14)<<26 | 3<<21 | 0<<16 | 42
	c, _ := newTestCPU([]uint32{instr})
	c.Step()
	if c.regs.GPR[3] != 42 {
		t.Fatalf("r3 = %d, want 42", c.regs.GPR[3])
	}
}

func TestAddSumsTwoRegisters(t *testing.T) {
	// addi r1,r0,10 ; addi r2,r0,32 ; add r3,r1,r2 (ext31=266)
	addi1 := uint32(14)<<26 | 1<<21 | 0<<16 | 10
	addi2 := uint32(14)<<26 | 2<<21 | 0<<16 | 32
	add3 := uint32(31)<<26 | 3<<21 | 1<<16 | 2<<11 | 266<<1
	c, _ := newTestCPU([]uint32{addi1, addi2, add3})
	c.Step()
	c.Step()
	c.Step()
	if c.regs.GPR[3] != 42 {
		t.Fatalf("r3 = %d, want 42", c.regs.GPR[3])
	}
}

func TestStoreAndLoadWord(t *testing.T) {
	// addi r1,r0,0x1234 ; stw r1,0x100(r0) ; lwz r2,0x100(r0)
	addi1 := uint32(14)<<26 | 1<<21 | 0<<16 | 0x1234
	stw := uint32(36)<<26 | 1<<21 | 0<<16 | 0x100
	lwz := uint32(32)<<26 | 2<<21 | 0<<16 | 0x100
	c, _ := newTestCPU([]uint32{addi1, stw, lwz})
	c.Step()
	c.Step()
	c.Step()
	if c.regs.GPR[2] != 0x1234 {
		t.Fatalf("r2 = %#08x, want 0x1234", c.regs.GPR[2])
	}
}

func TestUnconditionalBranch(t *testing.T) {
	// b +8 (skip the next instruction) ; addi r1,r0,99 (skipped) ; addi r1,r0,1
	b := uint32(18)<<26 | uint32(8&0x03FFFFFC)
	addiSkipped := uint32(14)<<26 | 1<<21 | 0<<16 | 99
	addiTarget := uint32(14)<<26 | 1<<21 | 0<<16 | 1
	c, _ := newTestCPU([]uint32{b, addiSkipped, addiTarget})
	c.Step()
	if c.PC() != 8 {
		t.Fatalf("PC after branch = %d, want 8", c.PC())
	}
	c.Step()
	if c.regs.GPR[1] != 1 {
		t.Fatalf("r1 = %d, want 1 (branch should have skipped the 99 load)", c.regs.GPR[1])
	}
}

func TestBranchAndLinkSetsLR(t *testing.T) {
	// bl +8
	bl := uint32(18)<<26 | uint32(8&0x03FFFFFC) | 1 // LK bit set
	c, _ := newTestCPU([]uint32{bl})
	c.Step()
	if c.regs.LR != 4 {
		t.Fatalf("LR = %d, want 4 (return address)", c.regs.LR)
	}
}

func TestBranchToLinkRegister(t *testing.T) {
	// mtlr r1 (after loading r1=0x100) ; bclr always
	addi1 := uint32(14)<<26 | 1<<21 | 0<<16 | 0x100
	mtlr := uint32(31)<<26 | 1<<21 | 8<<16 | 0<<11 | 467<<1
	bclr := uint32(19)<<26 | 20<<21 | 0<<16 | 0<<11 | 16<<1 // BO=20 (always branch)
	c, _ := newTestCPU([]uint32{addi1, mtlr, bclr})
	c.Step()
	c.Step()
	c.Step()
	if c.PC() != 0x100 {
		t.Fatalf("PC after bclr = %#08x, want 0x100", c.PC())
	}
}

func TestCompareSetsConditionRegister(t *testing.T) {
	// addi r1,r0,5 ; cmpi 0,0,r1,5 (BF=0, L=0, rA=1, SIMM=5)
	addi1 := uint32(14)<<26 | 1<<21 | 0<<16 | 5
	cmpi := uint32(11)<<26 | 0<<21 | 1<<16 | 5
	c, _ := newTestCPU([]uint32{addi1, cmpi})
	c.Step()
	c.Step()
	if c.regs.CR>>28 != 0x2 { // EQ bit set
		t.Fatalf("CR field0 = %#x, want 0x2 (equal)", c.regs.CR>>28)
	}
}
