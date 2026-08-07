package memory

// cartridge is a flat, unbanked ColecoVision ROM image, mirrored to
// fill whatever space it doesn't occupy - the (rare) bank-switched
// carts aren't implemented.
type cartridge struct {
	rom []byte
}

func newCartridge(rom []byte) cartridge { return cartridge{rom: rom} }

func (c *cartridge) read(addr uint16) byte {
	if len(c.rom) == 0 {
		return 0xFF
	}
	return c.rom[int(addr)%len(c.rom)]
}
