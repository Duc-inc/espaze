// mmu.go implements Block Address Translation (BAT): the coarse-grain
// address translation real PowerPC 6xx/7xx-family CPUs (including
// Gekko/Broadway) use to map large, fixed-size blocks of effective
// (virtual) address space onto physical RAM - notably how real
// GameCube software distinguishes the cached view of MEM1
// (0x80000000+) from the uncached view of the same physical RAM
// (0xC0000000+). This project implements only BAT translation, not
// full page-table-based paging (segment registers, the hashed page
// table, TLB) - real IPL/OS code typically covers all of usable RAM
// with a handful of BAT entries anyway, so this covers the common
// case without the much larger undertaking full paging would be.
// Instruction and data BATs are also not modeled separately: both
// read and write paths consult the same 8 entries, a simplification
// real hardware doesn't make (IBAT only applies to instruction
// fetches, DBAT only to data accesses).
package powerpc

// bat is one decoded Block Address Translation entry: a fixed-size
// block of effective address space (starting at effective<<17, a
// multiple of 128KB, real hardware's own BAT granularity) mapped onto
// physical address real<<17.
type bat struct {
	effective uint32 // BEPI: top 15 bits of the effective base address
	length    uint32 // BL: 11-bit block-length field, block size = (length+1)*128KB
	real      uint32 // BRPN: top 15 bits of the physical base address
	valid     bool   // Vs or Vp set - this project doesn't distinguish supervisor/user mode
}

func decodeBATUpper(v uint32) (effective, length uint32, valid bool) {
	effective = (v >> 17) & 0x7FFF
	length = (v >> 2) & 0x7FF
	valid = v&0x3 != 0 // Vs (bit 30) or Vp (bit 31)
	return
}

func encodeBATUpper(b bat) uint32 {
	v := b.effective<<17 | b.length<<2
	if b.valid {
		v |= 0x3
	}
	return v
}

func decodeBATLower(v uint32) (real uint32) {
	return (v >> 17) & 0x7FFF
}

func encodeBATLower(b bat) uint32 {
	return b.real << 17
}

// translate maps an effective address to a physical one using the
// first matching BAT entry, or returns addr unchanged if none match -
// this project has no page tables to fall back to, so an address no
// BAT covers is treated as already physical (matching this project's
// behavior before BATs existed).
func (c *CPU) translate(addr uint32) uint32 {
	for _, b := range c.regs.bats {
		if !b.valid {
			continue
		}
		size := (b.length + 1) * 128 * 1024
		base := b.effective << 17
		if addr >= base && addr-base < size {
			return b.real<<17 + (addr - base)
		}
	}
	return addr
}

// SetBAT configures one of the 8 BAT entries (0-3 = IBAT0-3, 4-7 =
// DBAT0-3) directly, mapping sizeBytes of effective address space
// starting at effectiveBase onto physical memory starting at
// realBase - this project's own stand-in for what real IPL firmware
// would otherwise set up via mtspr before jumping into game code,
// the same role SetPC/SetGPR already play for other initial state. An
// out-of-range index is ignored; sizeBytes is rounded down to the
// nearest 128KB (real hardware's own BAT block granularity).
func (c *CPU) SetBAT(index int, effectiveBase, sizeBytes, realBase uint32) {
	if index < 0 || index >= len(c.regs.bats) {
		return
	}
	c.regs.bats[index] = bat{
		effective: effectiveBase >> 17,
		length:    sizeBytes/(128*1024) - 1,
		real:      realBase >> 17,
		valid:     true,
	}
}

func (c *CPU) read8(addr uint32) byte    { return c.bus.Read8(c.translate(addr)) }
func (c *CPU) read16(addr uint32) uint16 { return c.bus.Read16(c.translate(addr)) }
func (c *CPU) read32(addr uint32) uint32 { return c.bus.Read32(c.translate(addr)) }

func (c *CPU) write8(addr uint32, v byte)    { c.bus.Write8(c.translate(addr), v) }
func (c *CPU) write16(addr uint32, v uint16) { c.bus.Write16(c.translate(addr), v) }
func (c *CPU) write32(addr uint32, v uint32) { c.bus.Write32(c.translate(addr), v) }
