package memory

// cartridge is a flat GBA ROM image (up to 32MB), mirrored to fill
// whatever space it doesn't occupy - the various bank-switching/
// flash-based cartridges some later/larger games use aren't modeled.
type cartridge struct {
	rom []byte
}

func newCartridge(rom []byte) cartridge { return cartridge{rom: rom} }

func (c *cartridge) read8(addr uint32) byte {
	if len(c.rom) == 0 {
		return 0xFF
	}
	return c.rom[int(addr)%len(c.rom)]
}

func (c *cartridge) read16(addr uint32) uint16 {
	return uint16(c.read8(addr)) | uint16(c.read8(addr+1))<<8
}

func (c *cartridge) read32(addr uint32) uint32 {
	return uint32(c.read16(addr)) | uint32(c.read16(addr+2))<<16
}

// sram is a simple 32KB battery-backed save area - real SRAM/Flash/
// EEPROM auto-detection and the larger Flash/EEPROM save types aren't
// implemented, just the simplest (and most common on early carts) SRAM case.
type sram struct {
	data [0x8000]byte
}

func (s *sram) read(addr uint32) byte     { return s.data[addr&0x7FFF] }
func (s *sram) write(addr uint32, v byte) { s.data[addr&0x7FFF] = v }
