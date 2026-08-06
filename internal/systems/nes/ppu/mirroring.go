package ppu

// nametableBank resolves which of the PPU's four 1KB internal banks
// backs a given nametable index (0-3), according to the cartridge's
// mirroring. Real hardware only wires up two physical 1KB chips unless a
// four-screen cartridge supplies the other two - but the PPU here always
// keeps four banks (see PPU.nametables), so every mode is just a matter
// of which banks alias each other, with no special-casing.
func nametableBank(mode MirrorMode, table int) int {
	switch mode {
	case MirrorHorizontal:
		return table / 2
	case MirrorVertical:
		return table % 2
	case MirrorSingleScreenLow:
		return 0
	case MirrorSingleScreenHigh:
		return 1
	default: // MirrorFourScreen
		return table
	}
}
