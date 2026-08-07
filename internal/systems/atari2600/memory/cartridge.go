package memory

// cartridge is a flat, unbanked 2KB or 4KB Atari 2600 ROM image - the
// format the overwhelming majority of early-era cartridges use. Larger
// bank-switched carts (the various "F8"/"F6" schemes later games use)
// aren't implemented; a too-small image mirrors to fill 4KB, matching
// how 2KB carts actually appear twice in the CPU's address window on
// real hardware.
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
