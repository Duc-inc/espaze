package cpu

// Fixed physical addresses ST0/ST1/ST2 write to directly, bypassing
// the MMU entirely - real hardware's shortcut for touching the VDC's
// ports without dedicating one of the 8 MPR pages to it. These match
// the physical layout this project's own pcengine/memory package
// decodes (see its own doc comment for why that layout is this
// project's best-effort reconstruction rather than an independently
// verified real hardware map).
const (
	vdcRegSelectAddr = 0x1FE000
	vdcDataLoAddr    = 0x1FE002
	vdcDataHiAddr    = 0x1FE003
)

func init() {
	setOp(0x53, func(c *CPU) int { // TAM #imm: A -> selected MPR page(s)
		mask := c.fetchByte()
		c.mmu.writePages(mask, c.regs.A)
		return 5
	})
	setOp(0x43, func(c *CPU) int { // TMA #imm: selected MPR page -> A
		mask := c.fetchByte()
		c.regs.A = c.mmu.readPage(mask)
		return 4
	})

	setOp(0x03, func(c *CPU) int { c.bus.Write(vdcRegSelectAddr, c.fetchByte()); return 5 })
	setOp(0x13, func(c *CPU) int { c.bus.Write(vdcDataLoAddr, c.fetchByte()); return 5 })
	setOp(0x23, func(c *CPU) int { c.bus.Write(vdcDataHiAddr, c.fetchByte()); return 5 })

	// CSH/CSL (7.16MHz/1.79MHz clock speed select) and SET (an obscure
	// addressing-mode modifier for the instruction immediately
	// following it) aren't modeled - this project always runs the CPU
	// at a single fixed rate, and no game this project targets depends
	// on SET's effect for core gameplay logic.
	setOp(0xD4, func(c *CPU) int { return 3 })
	setOp(0x54, func(c *CPU) int { return 3 })
	setOp(0xF4, func(c *CPU) int { return 2 })
}
