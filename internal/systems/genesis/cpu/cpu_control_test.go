package cpu

import "testing"

func TestBraAndBsrRts(t *testing.T) {
	// BSR +2 (relative to the address right after the opcode word,
	// $1002 - landing on $1004) ; (return here) MOVEQ #1,D0 ; at $1004: RTS
	c, bus := newTestCPU([]byte{0x61, 0x02, 0x70, 0x01})
	bus.mem[0x1004] = 0x4E
	bus.mem[0x1005] = 0x75 // RTS

	c.Step() // BSR
	if c.PC() != 0x1004 {
		t.Fatalf("PC after BSR = %#06x, want 0x1004", c.PC())
	}
	c.Step() // RTS
	if c.PC() != 0x1002 {
		t.Fatalf("PC after RTS = %#06x, want 0x1002", c.PC())
	}
	c.Step() // MOVEQ #1,D0
	if c.regs.D[0] != 1 {
		t.Fatalf("D0 = %d, want 1", c.regs.D[0])
	}
}

func TestDbccLoopsUntilCounterExhausted(t *testing.T) {
	// MOVEQ #2,D0 ; DBRA D0,<branch to itself> (loops in place 3 times)
	// Displacement is relative to the address right after the DBRA
	// opcode word (i.e. the displacement word's own address, $1004);
	// -2 brings that back to $1002, the DBRA opcode itself.
	c, _ := newTestCPU([]byte{0x70, 0x02, 0x51, 0xC8, 0xFF, 0xFE})
	c.Step() // MOVEQ #2,D0

	iterations := 0
	for i := 0; i < 10; i++ {
		beforePC := c.PC()
		c.Step()
		iterations++
		if c.PC() == beforePC+4 { // fell through, loop's over
			break
		}
	}
	if int16(c.regs.D[0]) != -1 {
		t.Fatalf("D0 after loop = %d, want -1 (decremented past 0)", int16(c.regs.D[0]))
	}
	if iterations != 3 {
		t.Fatalf("loop ran %d times, want 3", iterations)
	}
}

func TestSccSetsAllBitsOnTrueCondition(t *testing.T) {
	// MOVEQ #0,D0 ; MOVEQ #0,D1 ; CMP.L D0,D1 ; SEQ D2
	c, _ := newTestCPU([]byte{
		0x70, 0x00,
		0x72, 0x00,
		0xB2, 0x80,
		0x57, 0xC2,
	})
	c.Step()
	c.Step()
	c.Step() // CMP.L D0,D1 -> equal, Z set
	c.Step() // SEQ D2
	if byte(c.regs.D[2]) != 0xFF {
		t.Fatalf("D2 low byte = %#02x, want 0xFF (condition true)", byte(c.regs.D[2]))
	}
}

func TestMovemPredecrementThenPostincrementRoundTrips(t *testing.T) {
	// MOVEM.L D0-D1,-(A7) ; MOVEQ #0,D0 ; MOVEQ #0,D1 ; MOVEM.L (A7)+,D0-D1
	c, _ := newTestCPU([]byte{
		0x48, 0xE7, 0xC0, 0x00, // MOVEM.L D0/D1,-(A7)  (mask bits for D0,D1 in predecrement order)
		0x70, 0x00,
		0x72, 0x00,
		0x4C, 0xDF, 0x00, 0x03, // MOVEM.L (A7)+,D0/D1
	})
	c.regs.D[0] = 0x11111111
	c.regs.D[1] = 0x22222222
	sp := c.regs.A[7]

	c.Step() // store
	if c.regs.A[7] != sp-8 {
		t.Fatalf("SP after store = %#06x, want %#06x", c.regs.A[7], sp-8)
	}
	c.Step() // clobber D0
	c.Step() // clobber D1
	c.Step() // reload

	if c.regs.D[0] != 0x11111111 || c.regs.D[1] != 0x22222222 {
		t.Fatalf("D0,D1 = %#010x,%#010x, want restored 0x11111111,0x22222222", c.regs.D[0], c.regs.D[1])
	}
	if c.regs.A[7] != sp {
		t.Fatalf("SP after reload = %#06x, want restored %#06x", c.regs.A[7], sp)
	}
}

func TestSwapExchangesHalves(t *testing.T) {
	c, _ := newTestCPU([]byte{0x48, 0x40}) // SWAP D0
	c.regs.D[0] = 0x12345678
	c.Step()
	if c.regs.D[0] != 0x56781234 {
		t.Fatalf("D0 = %#010x, want 0x56781234", c.regs.D[0])
	}
}

func TestExgSwapsFullRegisters(t *testing.T) {
	c, _ := newTestCPU([]byte{0xC1, 0x41}) // EXG D0,D1
	c.regs.D[0], c.regs.D[1] = 0xAAAA, 0xBBBB
	c.Step()
	if c.regs.D[0] != 0xBBBB || c.regs.D[1] != 0xAAAA {
		t.Fatalf("D0,D1 = %#x,%#x, want swapped 0xBBBB,0xAAAA", c.regs.D[0], c.regs.D[1])
	}
}

func TestMuluAndDivu(t *testing.T) {
	// MOVEQ #10,D0 ; MULU.W #20,D0 ; DIVU.W #7,D0
	c, _ := newTestCPU([]byte{
		0x70, 0x0A,
		0xC0, 0xFC, 0x00, 0x14,
		0x80, 0xFC, 0x00, 0x07,
	})
	c.Step()
	c.Step()
	if c.regs.D[0] != 200 {
		t.Fatalf("D0 after MULU = %d, want 200", c.regs.D[0])
	}
	c.Step()
	if uint16(c.regs.D[0]) != 28 { // quotient in low word
		t.Fatalf("DIVU quotient = %d, want 28", uint16(c.regs.D[0]))
	}
	if uint16(c.regs.D[0]>>16) != 4 { // remainder in high word
		t.Fatalf("DIVU remainder = %d, want 4", uint16(c.regs.D[0]>>16))
	}
}

func TestIRQServicedThroughAutovector(t *testing.T) {
	c, bus := newTestCPU(nil)
	bus.Write32((24+2)*4, 0x2000) // autovector for level 2 lives at vector*4
	c.regs.SR &^= srIPMask        // unmask all interrupt levels

	c.TriggerIRQ(2)
	c.Step()
	if c.PC() != 0x2000 {
		t.Fatalf("PC after IRQ = %#06x, want 0x2000", c.PC())
	}
	if c.regs.interruptMask() != 2 {
		t.Fatalf("interrupt mask after service = %d, want 2", c.regs.interruptMask())
	}
}

func TestTrapVectorsCorrectly(t *testing.T) {
	c, bus := newTestCPU([]byte{0x4E, 0x40}) // TRAP #0
	bus.Write32(vectorTrap0*4, 0x4000)
	c.Step()
	if c.PC() != 0x4000 {
		t.Fatalf("PC after TRAP #0 = %#06x, want 0x4000", c.PC())
	}
}
