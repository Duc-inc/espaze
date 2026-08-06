package ppu

// masterPalette maps the PPU's 64 possible color indices to the sRGB
// approximation of its NTSC composite output that emulators
// conventionally use (the NES has no concept of RGB internally).
var masterPalette = [64][3]byte{
	{84, 84, 84}, {0, 30, 116}, {8, 16, 144}, {48, 0, 136}, {68, 0, 100}, {92, 0, 48}, {84, 4, 0}, {60, 24, 0},
	{32, 42, 0}, {8, 58, 0}, {0, 64, 0}, {0, 60, 0}, {0, 50, 60}, {0, 0, 0}, {0, 0, 0}, {0, 0, 0},
	{152, 150, 152}, {8, 76, 196}, {48, 50, 236}, {92, 30, 228}, {136, 20, 176}, {160, 20, 100}, {152, 34, 32}, {120, 60, 0},
	{84, 90, 0}, {40, 114, 0}, {8, 124, 0}, {0, 118, 40}, {0, 102, 120}, {0, 0, 0}, {0, 0, 0}, {0, 0, 0},
	{236, 238, 236}, {76, 154, 236}, {120, 124, 236}, {176, 98, 236}, {228, 84, 236}, {236, 88, 180}, {236, 106, 100}, {212, 136, 32},
	{160, 170, 0}, {116, 196, 0}, {76, 208, 32}, {56, 204, 108}, {56, 180, 204}, {60, 60, 60}, {0, 0, 0}, {0, 0, 0},
	{236, 238, 236}, {168, 204, 236}, {188, 188, 236}, {212, 178, 236}, {236, 174, 236}, {236, 174, 212}, {236, 180, 176}, {228, 196, 144},
	{204, 210, 120}, {180, 222, 120}, {168, 226, 144}, {152, 226, 180}, {160, 214, 228}, {160, 162, 160}, {0, 0, 0}, {0, 0, 0},
}

// paletteRAM holds the PPU's 32-byte palette (one universal background
// color plus four 3-color background sub-palettes, then four 3-color
// sprite sub-palettes).
type paletteRAM struct {
	data [32]byte
}

// index applies the hardware's mirroring quirk: the sprite palette's
// four "transparent" slots ($3F10/$14/$18/$1C) actually read/write the
// background palette's equivalent slots instead of having their own.
func (p *paletteRAM) index(addr uint16) uint16 {
	addr &= 0x1F
	if addr >= 0x10 && addr%4 == 0 {
		addr -= 0x10
	}
	return addr
}

func (p *paletteRAM) read(addr uint16) byte     { return p.data[p.index(addr)] }
func (p *paletteRAM) write(addr uint16, v byte) { p.data[p.index(addr)] = v & 0x3F }
