package cpu

// CPU is the Sharp LR35902 instruction interpreter. It has no boot ROM:
// Reset() pokes in the exact register/IO state real hardware reaches
// right after the Nintendo logo boot sequence finishes, so cartridges
// boot straight into their own code the same way they would after it.
type CPU struct {
	regs registers
	bus  Bus

	ime     bool
	eiDelay int // >0 while an EI's effect is still pending (see Step)
	halted  bool
	stopped bool
}

// New wires a CPU to the bus it will read instructions and data from.
func New(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset sets registers and I/O to the well-known post-boot-ROM state.
func (c *CPU) Reset() {
	c.regs = registers{SP: 0xFFFE, PC: 0x0100}
	c.regs.SetAF(0x01B0)
	c.regs.SetBC(0x0013)
	c.regs.SetDE(0x00D8)
	c.regs.SetHL(0x014D)
	c.ime, c.eiDelay, c.halted, c.stopped = false, 0, false, false

	for addr, v := range postBootIO {
		c.bus.Write(addr, v)
	}
}

// postBootIO are the I/O register values real DMG hardware leaves behind
// once its internal boot ROM hands control to the cartridge.
var postBootIO = map[uint16]byte{
	0xFF05: 0x00, 0xFF06: 0x00, 0xFF07: 0x00,
	0xFF40: 0x91, 0xFF42: 0x00, 0xFF43: 0x00, 0xFF45: 0x00,
	0xFF47: 0xFC, 0xFF48: 0xFF, 0xFF49: 0xFF, 0xFF4A: 0x00, 0xFF4B: 0x00,
	0xFFFF: 0x00,
}

// PC exposes the program counter, mainly for tests.
func (c *CPU) PC() uint16 { return c.regs.PC }

// Step executes exactly one instruction (or services a pending interrupt,
// or idles if halted/stopped) and returns how many T-cycles it took.
func (c *CPU) Step() int {
	if c.eiDelay > 0 {
		c.eiDelay--
		if c.eiDelay == 0 {
			c.ime = true
		}
	}

	if c.stopped {
		return 4
	}

	if cycles := c.checkInterrupts(); cycles > 0 {
		return cycles
	}

	if c.halted {
		return 4
	}

	opcode := c.fetch8()
	if opcode == 0xCB {
		return c.executeCB(c.fetch8())
	}
	return c.execute(opcode)
}

func (c *CPU) fetch8() byte {
	v := c.bus.Read(c.regs.PC)
	c.regs.PC++
	return v
}

func (c *CPU) fetch16() uint16 {
	lo := uint16(c.fetch8())
	hi := uint16(c.fetch8())
	return hi<<8 | lo
}

func (c *CPU) push16(v uint16) {
	c.regs.SP--
	c.bus.Write(c.regs.SP, byte(v>>8))
	c.regs.SP--
	c.bus.Write(c.regs.SP, byte(v))
}

func (c *CPU) pop16() uint16 {
	lo := uint16(c.bus.Read(c.regs.SP))
	c.regs.SP++
	hi := uint16(c.bus.Read(c.regs.SP))
	c.regs.SP++
	return hi<<8 | lo
}

// readR8/writeR8 implement the standard 3-bit register-index encoding
// used throughout the opcode table: B,C,D,E,H,L,(HL),A.
func (c *CPU) readR8(idx byte) byte {
	switch idx {
	case 0:
		return c.regs.B
	case 1:
		return c.regs.C
	case 2:
		return c.regs.D
	case 3:
		return c.regs.E
	case 4:
		return c.regs.H
	case 5:
		return c.regs.L
	case 6:
		return c.bus.Read(c.regs.HL())
	default:
		return c.regs.A
	}
}

func (c *CPU) writeR8(idx byte, v byte) {
	switch idx {
	case 0:
		c.regs.B = v
	case 1:
		c.regs.C = v
	case 2:
		c.regs.D = v
	case 3:
		c.regs.E = v
	case 4:
		c.regs.H = v
	case 5:
		c.regs.L = v
	case 6:
		c.bus.Write(c.regs.HL(), v)
	default:
		c.regs.A = v
	}
}

// readR16/writeR16 cover the {BC,DE,HL,SP} grouping (LD rr,nn / INC rr /
// DEC rr / ADD HL,rr).
func (c *CPU) readR16(idx byte) uint16 {
	switch idx {
	case 0:
		return c.regs.BC()
	case 1:
		return c.regs.DE()
	case 2:
		return c.regs.HL()
	default:
		return c.regs.SP
	}
}

func (c *CPU) writeR16(idx byte, v uint16) {
	switch idx {
	case 0:
		c.regs.SetBC(v)
	case 1:
		c.regs.SetDE(v)
	case 2:
		c.regs.SetHL(v)
	default:
		c.regs.SP = v
	}
}

// readR16Stack/writeR16Stack cover the {BC,DE,HL,AF} grouping used by
// PUSH/POP.
func (c *CPU) readR16Stack(idx byte) uint16 {
	if idx == 3 {
		return c.regs.AF()
	}
	return c.readR16(idx)
}

func (c *CPU) writeR16Stack(idx byte, v uint16) {
	if idx == 3 {
		c.regs.SetAF(v)
		return
	}
	c.writeR16(idx, v)
}
