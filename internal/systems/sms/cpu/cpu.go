package cpu

// CPU is a from-scratch Zilog Z80 interpreter, decoded via its
// documented bit-field structure (x/y/z/p/q, see decode.go) rather than
// a flat per-opcode table - the Z80's main opcode space is regular
// enough that this stays both compact and easy to check against the
// official encoding tables.
type CPU struct {
	regs registers
	bus  Bus
	io   IOBus

	halted bool

	pendingNMI bool
	pendingINT bool
	intData    byte // the data byte a device would put on the bus for a real IM 2/0 interrupt ack

	eiDelay bool // DI/EI take effect after the *next* instruction, not immediately
}

// New wires a CPU to its memory and I/O buses.
func New(bus Bus, io IOBus) *CPU {
	c := &CPU{bus: bus, io: io}
	c.Reset()
	return c
}

// Reset matches the Z80's own reset state: PC/interrupt state cleared,
// SP and other registers left as real silicon leaves them (0xFFFF and
// undefined respectively - callers relying on undefined register
// contents at boot don't exist, so zero is as good as anything else).
func (c *CPU) Reset() {
	c.regs = registers{SP: 0xFFFF, I: 0, R: 0}
	c.halted = false
	c.pendingNMI, c.pendingINT = false, false
	c.eiDelay = false
}

// PC exposes the program counter, mainly for tests.
func (c *CPU) PC() uint16 { return c.regs.PC }

// TriggerNMI/TriggerInterrupt latch an interrupt line the VDP can
// raise; TriggerInterrupt's data byte only matters in IM 2 (SMS
// hardware ties it to 0xFF via a pull-up, so that's the only value this
// project ever needs to pass).
func (c *CPU) TriggerNMI()                { c.pendingNMI = true }
func (c *CPU) TriggerInterrupt(data byte) { c.pendingINT, c.intData = true, data }

func (c *CPU) fetchByte() byte {
	v := c.bus.Read(c.regs.PC)
	c.regs.PC++
	c.regs.R = (c.regs.R & 0x80) | ((c.regs.R + 1) & 0x7F)
	return v
}

func (c *CPU) fetch16() uint16 {
	lo := uint16(c.fetchByte())
	hi := uint16(c.fetchByte())
	return hi<<8 | lo
}

func (c *CPU) read16(addr uint16) uint16 {
	lo := uint16(c.bus.Read(addr))
	hi := uint16(c.bus.Read(addr + 1))
	return hi<<8 | lo
}

func (c *CPU) write16(addr uint16, v uint16) {
	c.bus.Write(addr, byte(v))
	c.bus.Write(addr+1, byte(v>>8))
}

func (c *CPU) push(v uint16) {
	c.regs.SP -= 2
	c.write16(c.regs.SP, v)
}

func (c *CPU) pop() uint16 {
	v := c.read16(c.regs.SP)
	c.regs.SP += 2
	return v
}

// Step services any pending interrupt, then executes exactly one
// instruction, returning how many T-states it took.
func (c *CPU) Step() int {
	if c.pendingNMI {
		c.pendingNMI = false
		c.halted = false
		c.regs.IFF2 = c.regs.IFF1
		c.regs.IFF1 = false
		c.push(c.regs.PC)
		c.regs.PC = 0x0066
		return 11
	}
	if c.pendingINT && c.regs.IFF1 && !c.eiDelay {
		c.pendingINT = false
		c.halted = false
		c.regs.IFF1, c.regs.IFF2 = false, false
		switch c.regs.IM {
		case 1:
			c.push(c.regs.PC)
			c.regs.PC = 0x0038
			return 13
		case 2:
			c.push(c.regs.PC)
			vector := uint16(c.regs.I)<<8 | uint16(c.intData)
			c.regs.PC = c.read16(vector)
			return 19
		default: // mode 0: SMS hardware always drives 0xFF (= RST 38h) here
			c.push(c.regs.PC)
			c.regs.PC = 0x0038
			return 13
		}
	}

	c.eiDelay = false

	if c.halted {
		c.regs.R = (c.regs.R & 0x80) | ((c.regs.R + 1) & 0x7F) // real hardware keeps refreshing memory while halted
		return 4
	}

	return c.decode()
}
