package ppu

// MirrorMode describes how the PPU's two hardware nametables are mapped
// onto the four logical nametable slots a game addresses - a property of
// the cartridge (fixed by wiring, or switchable by some mappers), not
// the PPU itself.
type MirrorMode int

const (
	MirrorHorizontal MirrorMode = iota
	MirrorVertical
	MirrorSingleScreenLow
	MirrorSingleScreenHigh
	MirrorFourScreen
)

// CartBus is the PPU's view of the cartridge: pattern table data (CHR
// ROM/RAM) and the current nametable mirroring, both of which live
// behind the mapper rather than the PPU itself, so a new mapper can
// change either without the PPU package knowing about it.
type CartBus interface {
	ReadCHR(addr uint16) byte
	WriteCHR(addr uint16, v byte)
	Mirroring() MirrorMode
}
