package ppu

// WriteVRAMAddrLow/WriteVRAMAddrHigh set the VRAM word address the
// data port reads/writes through.
func (p *PPU) WriteVRAMAddrLow(v byte)  { p.vramAddr = p.vramAddr&0xFF00 | uint16(v) }
func (p *PPU) WriteVRAMAddrHigh(v byte) { p.vramAddr = p.vramAddr&0x00FF | uint16(v)<<8 }

// WriteVRAMDataLow/WriteVRAMDataHigh implement the data port; a write
// to the high byte completes the word and auto-increments the address,
// matching real hardware's own two-byte-per-word protocol.
func (p *PPU) WriteVRAMDataLow(v byte) {
	p.vram[p.vramAddr] = p.vram[p.vramAddr]&0xFF00 | uint16(v)
}

func (p *PPU) WriteVRAMDataHigh(v byte) {
	p.vram[p.vramAddr] = p.vram[p.vramAddr]&0x00FF | uint16(v)<<8
	p.vramAddr++
}

func (p *PPU) ReadVRAMLow() byte  { return byte(p.vram[p.vramAddr]) }
func (p *PPU) ReadVRAMHigh() byte { v := byte(p.vram[p.vramAddr] >> 8); p.vramAddr++; return v }

// WriteCGRAMAddr/WriteCGRAMData implement CGRAM's byte-pair protocol:
// the address selects a color, and two consecutive data writes (low
// byte then high byte) supply its 15-bit BGR555 value.
func (p *PPU) WriteCGRAMAddr(v byte) { p.cgramAddr = v; p.cgramHigh = false }

func (p *PPU) WriteCGRAMData(v byte) {
	if !p.cgramHigh {
		p.cgram[p.cgramAddr] = p.cgram[p.cgramAddr]&0x7F00 | uint16(v)
		p.cgramHigh = true
		return
	}
	p.cgram[p.cgramAddr] = p.cgram[p.cgramAddr]&0x00FF | uint16(v&0x7F)<<8
	p.cgramHigh = false
	p.cgramAddr++
}

// WriteOAMAddr/WriteOAMData implement the sprite table's data port.
func (p *PPU) WriteOAMByte(addr uint16, v byte) { p.oam[addr&0x21F] = v }
func (p *PPU) ReadOAMByte(addr uint16) byte     { return p.oam[addr&0x21F] }

// WriteBGControl sets a background layer's bit depth (bit0: 0=2bpp,
// 1=4bpp) - real hardware's tilemap/character base address registers
// aren't modeled, since this project fixes the tile layout the same
// way its other tile-based PPUs do (see background.go).
func (p *PPU) WriteBGControl(layer int, v byte) { p.bgcnt[layer&0x03] = v }

func (p *PPU) WriteBGScrollH(layer int, v uint16) { p.bgHOfs[layer&0x03] = v }
func (p *PPU) WriteBGScrollV(layer int, v uint16) { p.bgVOfs[layer&0x03] = v }

// WriteMainScreen implements the main-screen layer/OBJ enable bits
// (bits0-3 = BG0-3, bit4 = OBJ).
func (p *PPU) WriteMainScreen(v byte) {
	p.bgMain = v & 0x0F
	p.objMain = v&0x10 != 0
}

func (p *PPU) bgEnabled(layer int) bool { return p.bgMain&(1<<uint(layer)) != 0 }
func (p *PPU) bg4bpp(layer int) bool    { return p.bgcnt[layer]&0x01 != 0 }
