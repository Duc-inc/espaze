package cpu

// bitOpcodes registers BTST/BCHG/BCLR/BSET, in both their dynamic
// (bit number in a data register) and static (bit number is the next
// instruction word) forms. Bit numbers wrap mod 32 against a data
// register destination, or mod 8 against a memory byte.
func bitOpcodes() []pattern {
	return []pattern{
		{mask: 0xF1C0, match: 0x0100, execute: makeBitOp(bitTest, true, false)},
		{mask: 0xF1C0, match: 0x0140, execute: makeBitOp(bitChange, false, false)},
		{mask: 0xF1C0, match: 0x0180, execute: makeBitOp(bitClear, false, false)},
		{mask: 0xF1C0, match: 0x01C0, execute: makeBitOp(bitSet, false, false)},
		{mask: 0xFFC0, match: 0x0800, execute: makeBitOp(bitTest, true, true)},
		{mask: 0xFFC0, match: 0x0840, execute: makeBitOp(bitChange, false, true)},
		{mask: 0xFFC0, match: 0x0880, execute: makeBitOp(bitClear, false, true)},
		{mask: 0xFFC0, match: 0x08C0, execute: makeBitOp(bitSet, false, true)},
	}
}

// bitOpKind mutates v's bit (if applicable) and returns the new value;
// the caller separately sets Z from the bit's *original* state.
type bitOpKind func(v uint32, bit uint32) uint32

func bitTest(v, bit uint32) uint32   { return v }
func bitChange(v, bit uint32) uint32 { return v ^ bit }
func bitClear(v, bit uint32) uint32  { return v &^ bit }
func bitSet(v, bit uint32) uint32    { return v | bit }

func makeBitOp(kind bitOpKind, readOnly bool, static bool) executeFunc {
	return func(c *CPU, opcode uint16) int {
		mode := byte((opcode >> 3) & 0x07)
		eaReg := byte(opcode & 0x07)

		var bitNum byte
		if static {
			bitNum = byte(c.fetchWord())
		} else {
			dReg := byte((opcode >> 9) & 0x07)
			bitNum = byte(c.regs.D[dReg])
		}

		if mode == 0 { // destination is a data register: 32-bit, mod 32
			bitNum &= 31
			bit := uint32(1) << bitNum
			v := c.regs.D[eaReg]
			c.regs.setFlag(FlagZ, v&bit == 0)
			if !readOnly {
				c.regs.D[eaReg] = kind(v, bit)
			}
			return 6
		}

		// destination is memory: a single byte, mod 8. BTST never writes
		// back - real hardware only issues a read cycle for it, which
		// matters for memory-mapped I/O registers a write could disturb.
		bitNum &= 7
		bit := uint32(1) << bitNum
		loc := c.resolveEA(mode, eaReg, sizeByte)
		v := c.readLocation(loc, sizeByte)
		c.regs.setFlag(FlagZ, v&bit == 0)
		if !readOnly {
			c.writeLocation(loc, sizeByte, kind(v, bit))
		}
		return 8
	}
}
