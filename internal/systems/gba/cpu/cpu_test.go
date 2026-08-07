package cpu

import "testing"

// testBus covers two small windows: low memory (exception vectors,
// $0-$FFFF) and the cartridge base ($08000000+), enough for these
// tests without allocating a full 32-bit address space.
type testBus struct {
	mem [0x20000]byte
}

func (b *testBus) idx(addr uint32) uint32 {
	if addr >= cartridgeEntry {
		return addr - cartridgeEntry + 0x10000
	}
	return addr
}

func (b *testBus) Read8(addr uint32) byte { return b.mem[b.idx(addr)] }
func (b *testBus) Read16(addr uint32) uint16 {
	i := b.idx(addr)
	return uint16(b.mem[i]) | uint16(b.mem[i+1])<<8
}
func (b *testBus) Read32(addr uint32) uint32 {
	i := b.idx(addr)
	return uint32(b.mem[i]) | uint32(b.mem[i+1])<<8 | uint32(b.mem[i+2])<<16 | uint32(b.mem[i+3])<<24
}
func (b *testBus) Write8(addr uint32, v byte) { b.mem[b.idx(addr)] = v }
func (b *testBus) Write16(addr uint32, v uint16) {
	i := b.idx(addr)
	b.mem[i], b.mem[i+1] = byte(v), byte(v>>8)
}
func (b *testBus) Write32(addr uint32, v uint32) {
	i := b.idx(addr)
	b.mem[i], b.mem[i+1], b.mem[i+2], b.mem[i+3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
}

func newTestCPU(armProgram []byte) (*CPU, *testBus) {
	bus := &testBus{}
	copy(bus.mem[bus.idx(cartridgeEntry):], armProgram)
	return New(bus), bus
}

func writeThumb(bus *testBus, addr uint32, program []uint16) {
	for i, w := range program {
		off := bus.idx(addr) + uint32(i*2)
		bus.mem[off], bus.mem[off+1] = byte(w), byte(w>>8)
	}
}

func TestResetStartsInARMStateAtCartridgeEntry(t *testing.T) {
	c, _ := newTestCPU(nil)
	if c.PC() != cartridgeEntry {
		t.Fatalf("PC = %#08x, want %#08x", c.PC(), cartridgeEntry)
	}
	if c.regs.thumb() {
		t.Fatal("expected ARM state at reset")
	}
}

func TestArmMovImmediate(t *testing.T) {
	c, _ := newTestCPU([]byte{0x01, 0x00, 0xA0, 0xE3}) // MOV R0,#1
	c.Step()
	if c.regs.R[0] != 1 {
		t.Fatalf("R0 = %d, want 1", c.regs.R[0])
	}
}

func TestBXSwitchesToThumbAndClearsBit0(t *testing.T) {
	c, _ := newTestCPU(nil)
	writeARM(c, cartridgeEntry, 0xE12FFF10) // BX R0
	c.regs.R[0] = cartridgeEntry + 0x100 + 1

	c.Step()
	if !c.regs.thumb() {
		t.Fatal("expected Thumb state after BX with bit0 set")
	}
	if c.PC() != cartridgeEntry+0x100 {
		t.Fatalf("PC after BX = %#08x, want %#08x", c.PC(), cartridgeEntry+0x100)
	}
}

func writeARM(c *CPU, addr uint32, word uint32) {
	c.bus.Write32(addr, word)
}

func TestThumbMovAndAdd(t *testing.T) {
	c, bus := newTestCPU(nil)
	c.regs.setFlag(FlagThumb, true)
	c.regs.R[15] = cartridgeEntry
	// MOV R0,#5 ; ADD R1,R0,#3
	writeThumb(bus, cartridgeEntry, []uint16{0x2005, 0x1CC1})
	c.Step()
	c.Step()
	if c.regs.R[1] != 8 {
		t.Fatalf("R1 = %d, want 8", c.regs.R[1])
	}
}

func TestThumbBranchLink(t *testing.T) {
	c, bus := newTestCPU(nil)
	c.regs.setFlag(FlagThumb, true)
	c.regs.R[15] = cartridgeEntry
	// BL with offsetHigh=0, offsetLow=2 halfwords: LR latches to
	// entry+4 (PC+4 convention) after the first halfword, then the
	// second adds offsetLow*2=4 more, landing on entry+8.
	writeThumb(bus, cartridgeEntry, []uint16{0xF000, 0xF802})
	c.Step() // first half: latches LR
	c.Step() // second half: jumps, sets LR to return address | 1
	if c.PC() != cartridgeEntry+8 {
		t.Fatalf("PC after BL = %#08x, want %#08x", c.PC(), cartridgeEntry+8)
	}
	if c.regs.R[14]&1 == 0 {
		t.Fatal("expected LR's Thumb bit to be set")
	}
}

func TestThumbPushPop(t *testing.T) {
	c, bus := newTestCPU(nil)
	c.regs.setFlag(FlagThumb, true)
	c.regs.R[15] = cartridgeEntry
	c.regs.R[13] = 0x1000 // within the test bus's low window
	c.regs.R[0] = 0xCAFEBABE
	// PUSH {R0} ; MOV R0,#0 ; POP {R0}
	writeThumb(bus, cartridgeEntry, []uint16{0xB401, 0x2000, 0xBC01})
	c.Step()
	c.Step()
	if c.regs.R[0] != 0 {
		t.Fatalf("R0 after clearing = %#08x, want 0", c.regs.R[0])
	}
	c.Step()
	if c.regs.R[0] != 0xCAFEBABE {
		t.Fatalf("R0 after pop = %#08x, want 0xCAFEBABE", c.regs.R[0])
	}
}

func TestArmDataProcessingImmediateAndFlags(t *testing.T) {
	c, _ := newTestCPU(nil)
	// SUBS R0,R0,#1 with R0=0 -> result=0xFFFFFFFF, borrow means Carry clear
	c.regs.R[0] = 0
	writeARM(c, cartridgeEntry, 0xE2500001)
	c.Step()
	if c.regs.R[0] != 0xFFFFFFFF {
		t.Fatalf("R0 = %#08x, want 0xFFFFFFFF", c.regs.R[0])
	}
	if c.regs.getFlag(FlagC) {
		t.Fatal("expected Carry clear (borrow occurred)")
	}
	if !c.regs.getFlag(FlagN) {
		t.Fatal("expected Negative set")
	}
}

func TestIRQServicedAndReturnsViaSPSR(t *testing.T) {
	c, _ := newTestCPU(nil)
	c.regs.setFlag(FlagIRQD, false)
	c.regs.R[15] = cartridgeEntry

	c.TriggerIRQ()
	c.Step() // services the IRQ, jumps to 0x18
	if c.PC() != 0x18 {
		t.Fatalf("PC after IRQ = %#08x, want 0x18", c.PC())
	}
	if c.regs.mode() != modeIRQ {
		t.Fatalf("mode after IRQ = %#x, want IRQ", c.regs.mode())
	}
}
