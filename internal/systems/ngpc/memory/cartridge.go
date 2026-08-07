package memory

// cartridge is a flat, unbanked NGPC ROM image, mirrored to fill
// whatever space it doesn't occupy.
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
