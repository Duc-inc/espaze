package cpu

// blockMode selects how the source/destination pointers evolve on
// each byte of a block-transfer instruction.
type blockMode int

const (
	modeTII blockMode = iota // src++, dst++
	modeTDD                  // src--, dst--
	modeTIA                  // src++, dst alternates between dst/dst+1 (VDC port ping-pong)
	modeTAI                  // src alternates between src/src+1, dst++
	modeTIN                  // src++, dst fixed (write a block to one I/O port)
)

// blockTransfer implements the HuC6280's TII/TDD/TIA/TAI/TIN family:
// each reads a 2-byte source, 2-byte destination, and 2-byte length
// from the instruction stream, then moves that many bytes in one
// atomic step (real hardware's own DMA-like execution, which likewise
// runs to completion without being interruptible).
func blockTransfer(c *CPU, mode blockMode) int {
	src := c.fetch16()
	dst := c.fetch16()
	length := c.fetch16()
	if length == 0 {
		length = 0x10000 - 1 // 0 means "as large as the 16-bit count can represent"
	}

	toggle := false
	for i := uint32(0); i < uint32(length); i++ {
		v := c.read(src)
		c.write(dst, v)

		switch mode {
		case modeTII:
			src++
			dst++
		case modeTDD:
			src--
			dst--
		case modeTIA:
			src++
			if toggle {
				dst--
			} else {
				dst++
			}
			toggle = !toggle
		case modeTAI:
			if toggle {
				src--
			} else {
				src++
			}
			toggle = !toggle
			dst++
		case modeTIN:
			src++
		}
	}
	return 17 + int(length)*6
}

func testBits(c *CPU, mask byte, l location) {
	v := c.readLoc(l)
	c.regs.setFlag(FlagZero, mask&v == 0)
	c.regs.setFlag(FlagOverflow, v&0x40 != 0)
	c.regs.setFlag(FlagNegative, v&0x80 != 0)
}

func init() {
	setOp(0x73, func(c *CPU) int { return blockTransfer(c, modeTII) })
	setOp(0xC3, func(c *CPU) int { return blockTransfer(c, modeTDD) })
	setOp(0xE3, func(c *CPU) int { return blockTransfer(c, modeTIA) })
	setOp(0xF3, func(c *CPU) int { return blockTransfer(c, modeTAI) })
	setOp(0xD3, func(c *CPU) int { return blockTransfer(c, modeTIN) })

	setOp(0x83, func(c *CPU) int { mask := c.fetchByte(); testBits(c, mask, c.zeroPage()); return 7 })
	setOp(0xA3, func(c *CPU) int { mask := c.fetchByte(); testBits(c, mask, c.zeroPageX()); return 7 })
	setOp(0x93, func(c *CPU) int { mask := c.fetchByte(); testBits(c, mask, c.absolute()); return 8 })
	setOp(0xB3, func(c *CPU) int { mask := c.fetchByte(); testBits(c, mask, c.absoluteX()); return 8 })
}
