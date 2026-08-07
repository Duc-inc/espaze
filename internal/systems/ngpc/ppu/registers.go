package ppu

// WriteControl implements this project's single display-control
// register: bit0 enables the background, bit1 enables sprites.
func (p *PPU) WriteControl(v byte) {
	p.bgEnable = v&0x01 != 0
	p.objEnable = v&0x02 != 0
}

func (p *PPU) WriteScrollX(v byte) { p.scrollX = v }
func (p *PPU) WriteScrollY(v byte) { p.scrollY = v }

// VRAM/sprite table/palette byte-level access.
func (p *PPU) ReadVRAM(addr uint32) byte     { return p.vram[addr&0x3FFF] }
func (p *PPU) WriteVRAM(addr uint32, v byte) { p.vram[addr&0x3FFF] = v }

func (p *PPU) ReadSprite(addr uint32) byte     { return p.sprites[addr&0xFF] }
func (p *PPU) WriteSprite(addr uint32, v byte) { p.sprites[addr&0xFF] = v }

func (p *PPU) WritePaletteLow(index byte, v byte) {
	p.palette[index&0x1F] = p.palette[index&0x1F]&0x0F00 | uint16(v)
}

func (p *PPU) WritePaletteHigh(index byte, v byte) {
	p.palette[index&0x1F] = p.palette[index&0x1F]&0x00FF | uint16(v&0x0F)<<8
}
