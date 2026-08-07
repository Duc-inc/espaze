package memory

// cartridge is a Genesis ROM image: on real hardware nearly always a
// flat, unbanked address space (unlike NES/SMS, bank-switching
// cartridges are the exception rather than the rule) mapped starting
// at $000000. Writes are dropped - SRAM-equipped carts exist but
// aren't modeled here.
type cartridge struct {
	rom []byte
}

func newCartridge(rom []byte) cartridge { return cartridge{rom: rom} }

func (c *cartridge) read8(addr uint32) byte {
	if int(addr) < len(c.rom) {
		return c.rom[addr]
	}
	return 0xFF
}

// read16/read32 are big-endian, matching the 68000's own byte order
// and the format Genesis ROM images are stored in.
func (c *cartridge) read16(addr uint32) uint16 {
	return uint16(c.read8(addr))<<8 | uint16(c.read8(addr+1))
}

func (c *cartridge) read32(addr uint32) uint32 {
	return uint32(c.read16(addr))<<16 | uint32(c.read16(addr+2))
}
