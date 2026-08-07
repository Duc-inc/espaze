package memory

// cartridge is a flat, unbanked HuCard ROM image, mirrored to fill
// out whatever physical space it doesn't occupy - bank-switched HuCards
// (a minority of the library) aren't implemented.
type cartridge struct {
	rom []byte
}

func newCartridge(rom []byte) cartridge { return cartridge{rom: rom} }

func (c *cartridge) read(addr uint32) byte {
	if len(c.rom) == 0 {
		return 0xFF
	}
	return c.rom[int(addr)%len(c.rom)]
}
