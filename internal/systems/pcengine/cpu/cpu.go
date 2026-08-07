// Package cpu implements the PC Engine's HuC6280 from scratch: a
// 65C02-family core (the base 6502 instruction set plus the 65C02's
// own additions - PHX/PLX/PHY/PLY, STZ, BRA, TRB/TSB, and zero-page
// indirect addressing) extended with the chip's own memory-mapping
// unit, block-transfer instructions, and built-in timer/interrupt
// controller. The HuC6280-specific extension opcodes are assigned the
// byte values commonly documented in PC Engine reference material;
// this project hasn't independently verified them against real
// hardware, so treat them as a best effort rather than a guaranteed
// match.
package cpu

// CPU is a from-scratch HuC6280 interpreter.
type CPU struct {
	regs registers
	mmu  mmu
	irq  irqTimer
	bus  Bus

	pendingNMI bool
}

// New wires a CPU to its physical bus and resets it.
func New(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// Reset loads every MPR page with 0 (mapping the CPU's entire logical
// space to the start of cartridge ROM) and reads PC from the logical
// $FFFE-$FFFF vector, which then lands on physical $1FFE-$1FFF -
// wherever a ROM padded/sized to at least 8KB would place its own
// reset vector, the same NES-style convention this project's other
// 6502-family cores use. Real hardware's actual boot sequence (a
// built-in system card sets up the MMU before jumping into the
// HuCard) isn't reproduced - this project has no BIOS, so a cartridge
// simply needs its own reset vector at the end of its first 8KB.
func (c *CPU) Reset() {
	c.regs = registers{S: 0xFD, P: FlagInterrupt | FlagUnused}
	c.mmu = mmu{}
	c.irq = irqTimer{}
	c.pendingNMI = false
	c.regs.PC = c.read16(0xFFFE)
}

// PC exposes the program counter, mainly for tests.
func (c *CPU) PC() uint16 { return c.regs.PC }

// TriggerIRQ2/TriggerIRQ1 latch the VDC's/an external device's
// interrupt line.
func (c *CPU) TriggerIRQ2() { c.irq.TriggerIRQ2() }
func (c *CPU) TriggerIRQ1() { c.irq.TriggerIRQ1() }

// TriggerNMI latches a non-maskable interrupt.
func (c *CPU) TriggerNMI() { c.pendingNMI = true }

// WriteTimerReload/WriteTimerControl/WriteIRQMask/ReadIRQStatus expose
// the built-in timer/interrupt controller's registers so the physical
// bus decode (in the top-level pcengine package) can dispatch writes
// to them, exactly like any other memory-mapped device.
func (c *CPU) WriteTimerReload(v byte)  { c.irq.writeTimerReload(v) }
func (c *CPU) WriteTimerControl(v byte) { c.irq.writeTimerControl(v) }
func (c *CPU) WriteIRQMask(v byte)      { c.irq.writeMask(v) }
func (c *CPU) ReadIRQStatus() byte      { return c.irq.readStatus() }

func (c *CPU) read(logical uint16) byte {
	return c.bus.Read(c.mmu.translate(logical))
}

func (c *CPU) write(logical uint16, v byte) {
	c.bus.Write(c.mmu.translate(logical), v)
}

func (c *CPU) read16(logical uint16) uint16 {
	lo := uint16(c.read(logical))
	hi := uint16(c.read(logical + 1))
	return lo | hi<<8
}

func (c *CPU) fetchByte() byte {
	v := c.read(c.regs.PC)
	c.regs.PC++
	return v
}

func (c *CPU) fetch16() uint16 {
	lo := uint16(c.fetchByte())
	hi := uint16(c.fetchByte())
	return lo | hi<<8
}

func (c *CPU) push(v byte) {
	c.write(0x2100|uint16(c.regs.S), v)
	c.regs.S--
}

func (c *CPU) pop() byte {
	c.regs.S++
	return c.read(0x2100 | uint16(c.regs.S))
}

func (c *CPU) push16(v uint16) {
	c.push(byte(v >> 8))
	c.push(byte(v))
}

func (c *CPU) pop16() uint16 {
	lo := uint16(c.pop())
	hi := uint16(c.pop())
	return lo | hi<<8
}

// Step services any pending interrupt (NMI takes priority over the
// three maskable IRQ sources), then executes exactly one instruction,
// returning how many clock cycles it took. The timer is advanced by
// the same amount before returning.
func (c *CPU) Step() int {
	var cycles int
	switch {
	case c.pendingNMI:
		c.pendingNMI = false
		cycles = c.serviceInterrupt(0xFFFC, false)
	case !c.regs.getFlag(FlagInterrupt):
		if vector, ok := c.irq.pendingVector(); ok {
			if vector == 0xFFFA {
				c.irq.acknowledgeTimer()
			} else if vector == 0xFFF6 {
				c.irq.pendingIRQ2 = false
			} else {
				c.irq.pendingIRQ1 = false
			}
			cycles = c.serviceInterrupt(vector, false)
		}
	}
	if cycles == 0 {
		opcode := c.fetchByte()
		entry := dispatchTable[opcode]
		if entry.execute == nil {
			cycles = 2 // undefined opcode: treated as a 2-cycle NOP
		} else {
			cycles = entry.execute(c)
		}
	}

	c.irq.step(cycles)
	return cycles
}

func (c *CPU) serviceInterrupt(vector uint16, brk bool) int {
	c.push16(c.regs.PC)
	flags := c.regs.P | FlagUnused
	if brk {
		flags |= FlagBreak
	}
	c.push(flags)
	c.regs.setFlag(FlagInterrupt, true)
	c.regs.setFlag(FlagDecimal, false) // HuC6280 clears D on interrupt entry, unlike NMOS 6502
	c.regs.PC = c.read16(vector)
	return 7
}
