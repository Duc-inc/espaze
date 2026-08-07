package cpu

import "testing"

func TestCBBitSetRes(t *testing.T) {
	c, _ := newTestCPU([]byte{0x3E, 0x00, 0xCB, 0xC7, 0xCB, 0x47, 0xCB, 0x87}) // LD A,0 ; SET 0,A ; BIT 0,A ; RES 0,A
	c.Step()                                                                   // LD A,0
	c.Step()                                                                   // SET 0,A
	if c.regs.A != 0x01 {
		t.Fatalf("A after SET 0,A = %#02x, want 0x01", c.regs.A)
	}
	c.Step() // BIT 0,A
	if c.regs.getFlag(FlagZ) {
		t.Fatal("BIT 0,A should find the bit set (Z clear)")
	}
	c.Step() // RES 0,A
	if c.regs.A != 0x00 {
		t.Fatalf("A after RES 0,A = %#02x, want 0x00", c.regs.A)
	}
}

func TestCBRotateLeftCircular(t *testing.T) {
	c, _ := newTestCPU([]byte{0x3E, 0x80, 0xCB, 0x07}) // LD A,0x80 ; RLC A
	c.Step()
	c.Step()
	if c.regs.A != 0x01 {
		t.Fatalf("A = %#02x, want 0x01 (bit 7 rotated into bit 0)", c.regs.A)
	}
	if !c.regs.getFlag(FlagC) {
		t.Fatal("expected Carry set from the rotated-out bit")
	}
}

func TestLDIRBlockCopy(t *testing.T) {
	// LDIR from 0x1000 (3 bytes) to 0x2000.
	c, bus := newTestCPU([]byte{0x21, 0x00, 0x10, 0x11, 0x00, 0x20, 0x01, 0x03, 0x00, 0xED, 0xB0})
	bus.mem[0x1000], bus.mem[0x1001], bus.mem[0x1002] = 0xAA, 0xBB, 0xCC

	c.Step() // LD HL
	c.Step() // LD DE
	c.Step() // LD BC
	for i := 0; i < 3; i++ {
		c.Step() // LDIR re-triggers itself (rewinds PC) once per byte, until BC == 0
	}
	if bus.mem[0x2000] != 0xAA || bus.mem[0x2001] != 0xBB || bus.mem[0x2002] != 0xCC {
		t.Fatalf("copied bytes = %#02x %#02x %#02x, want AA BB CC",
			bus.mem[0x2000], bus.mem[0x2001], bus.mem[0x2002])
	}
	if c.regs.BC() != 0 {
		t.Fatalf("BC after LDIR = %#04x, want 0", c.regs.BC())
	}
}

func TestCPIRFindsByte(t *testing.T) {
	// Search 4 bytes at 0x1000 for 0xCC.
	c, bus := newTestCPU([]byte{0x21, 0x00, 0x10, 0x01, 0x04, 0x00, 0x3E, 0xCC, 0xED, 0xB1})
	bus.mem[0x1000], bus.mem[0x1001], bus.mem[0x1002], bus.mem[0x1003] = 0xAA, 0xBB, 0xCC, 0xDD

	c.Step() // LD HL
	c.Step() // LD BC
	c.Step() // LD A,0xCC
	for i := 0; i < 3; i++ {
		c.Step() // CPIR re-triggers itself (rewinds PC) once per byte, until it matches or BC == 0
	}
	if !c.regs.getFlag(FlagZ) {
		t.Fatal("expected Zero: CPIR should have found the byte")
	}
	if c.regs.HL() != 0x1003 {
		t.Fatalf("HL after CPIR = %#04x, want 0x1003 (one past the match)", c.regs.HL())
	}
}

func TestIndexedLoadAndStore(t *testing.T) {
	// LD IX,0x1000 ; LD (IX+2),0x55 ; LD A,(IX+2)
	c, _ := newTestCPU([]byte{0xDD, 0x21, 0x00, 0x10, 0xDD, 0x36, 0x02, 0x55, 0xDD, 0x7E, 0x02})
	c.Step()
	c.Step()
	c.Step()
	if c.regs.A != 0x55 {
		t.Fatalf("A = %#02x, want 0x55", c.regs.A)
	}
}

func TestIndexedPassThroughForNonHLOpcodes(t *testing.T) {
	// DD-prefixed LD B,n doesn't touch HL/IX at all - must behave
	// identically to the unprefixed instruction.
	c, _ := newTestCPU([]byte{0xDD, 0x06, 0x77})
	c.Step()
	if c.regs.B != 0x77 {
		t.Fatalf("B = %#02x, want 0x77", c.regs.B)
	}
}

func TestIM1InterruptVectorsToRST38(t *testing.T) {
	c, bus := newTestCPU([]byte{0xFB, 0xED, 0x56}) // EI ; IM 1
	bus.mem[0x0038] = 0x00                         // NOP at the IM1 vector, just so Step() there is harmless

	c.Step() // EI (interrupts still masked for one more instruction)
	c.Step() // IM 1
	c.TriggerInterrupt(0xFF)

	c.Step() // should now service the interrupt instead of executing at PC
	if c.regs.PC != 0x0038 {
		t.Fatalf("PC after IM1 interrupt = %#04x, want 0x0038", c.regs.PC)
	}
	if c.regs.IFF1 {
		t.Fatal("IFF1 should be cleared while servicing an interrupt")
	}
}

func TestNMIVectorsTo0x0066(t *testing.T) {
	c, _ := newTestCPU(nil)
	c.TriggerNMI()
	c.Step()
	if c.regs.PC != 0x0066 {
		t.Fatalf("PC after NMI = %#04x, want 0x0066", c.regs.PC)
	}
}

func TestHaltWaitsForInterrupt(t *testing.T) {
	c, _ := newTestCPU([]byte{0x76}) // HALT
	c.Step()
	if !c.halted {
		t.Fatal("expected the CPU to be halted")
	}
	pc := c.regs.PC
	c.Step()
	if c.regs.PC != pc {
		t.Fatal("PC should not advance while halted")
	}
	c.TriggerNMI()
	c.Step()
	if c.halted {
		t.Fatal("NMI should wake the CPU from HALT")
	}
}
