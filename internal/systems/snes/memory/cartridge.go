package memory

// cartridge is a flat, unbanked SNES ROM image (LoROM/HiROM bank
// switching aren't modeled - every game's ROM is treated as one flat
// image mirrored to fill whatever space it doesn't occupy).
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
