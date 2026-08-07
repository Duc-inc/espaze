package ppu

// WriteDISPCNT/ReadDISPCNT implement the display control register.
func (p *PPU) WriteDISPCNT(v uint16) { p.dispcnt = v }
func (p *PPU) ReadDISPCNT() uint16   { return p.dispcnt }

// WriteBGCNT/WriteBGHOFS/WriteBGVOFS implement each background's
// control and scroll registers (bg is 0-3).
func (p *PPU) WriteBGCNT(bg int, v uint16)  { p.bgcnt[bg&0x03] = v }
func (p *PPU) WriteBGHOFS(bg int, v uint16) { p.bghofs[bg&0x03] = v & 0x01FF }
func (p *PPU) WriteBGVOFS(bg int, v uint16) { p.bgvofs[bg&0x03] = v & 0x01FF }

// VRAM/OAM/palette byte-level access, used by the memory bus for
// arbitrary byte/halfword/word reads and writes.
func (p *PPU) ReadVRAM8(addr uint32) byte     { return p.vram[addr&0x17FFF] }
func (p *PPU) WriteVRAM8(addr uint32, v byte) { p.vram[addr&0x17FFF] = v }
func (p *PPU) ReadOAM8(addr uint32) byte      { return p.oam[addr&0x3FF] }
func (p *PPU) WriteOAM8(addr uint32, v byte)  { p.oam[addr&0x3FF] = v }

func (p *PPU) ReadPalette8(addr uint32) byte {
	idx := (addr & 0x3FF) / 2
	c := p.palette[idx&0x1FF]
	if addr&1 == 0 {
		return byte(c)
	}
	return byte(c >> 8)
}

func (p *PPU) WritePalette16(addr uint32, v uint16) {
	idx := (addr & 0x3FF) / 2
	p.palette[idx&0x1FF] = v & 0x7FFF
}

func (p *PPU) vramWord(addr uint32) uint16 {
	i := addr & 0x17FFF
	return uint16(p.vram[i]) | uint16(p.vram[i+1])<<8
}

func (p *PPU) oamWord(addr uint32) uint16 {
	i := addr & 0x3FF
	return uint16(p.oam[i]) | uint16(p.oam[i+1])<<8
}
