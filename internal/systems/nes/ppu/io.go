package ppu

// ReadRegister implements CPU access to $2000-$2007 (mirrored every 8
// bytes through $3FFF - callers are expected to mask addr first).
func (p *PPU) ReadRegister(addr uint16) byte {
	switch addr % 8 {
	case 2: // PPUSTATUS
		v := p.status
		p.status &^= statusVBlank
		p.scroll.resetWriteToggle()
		return v
	case 4: // OAMDATA
		return p.oamMem.readByte(p.oamAddr)
	case 7: // PPUDATA
		return p.readData()
	default:
		return 0
	}
}

// WriteRegister implements CPU access to $2000-$2007.
func (p *PPU) WriteRegister(addr uint16, v byte) {
	switch addr % 8 {
	case 0: // PPUCTRL
		p.ctrl = v
		p.scroll.writeCtrlNametable(v & ctrlNametableMask)
	case 1: // PPUMASK
		p.mask = v
	case 3: // OAMADDR
		p.oamAddr = v
	case 4: // OAMDATA
		p.oamMem.writeByte(p.oamAddr, v)
		p.oamAddr++
	case 5: // PPUSCROLL
		p.scroll.writeScroll(v)
	case 6: // PPUADDR
		p.scroll.writeAddr(v)
	case 7: // PPUDATA
		p.writeData(v)
	}
}

func (p *PPU) vramIncrement() uint16 {
	if p.ctrl&ctrlVRAMIncrement != 0 {
		return 32
	}
	return 1
}

// readData implements $2007's read-buffering quirk: reads below the
// palette return the *previous* buffered byte, refilling the buffer
// with the newly read one - except palette reads, which return
// immediately and refill the buffer from the nametable data mirrored
// "underneath" the palette region instead.
func (p *PPU) readData() byte {
	addr := p.scroll.v & 0x3FFF
	var value byte
	if addr < 0x3F00 {
		value = p.dataBuffer
		p.dataBuffer = p.readVRAM(addr)
	} else {
		value = p.palette.read(addr)
		p.dataBuffer = p.readVRAM(addr - 0x1000)
	}
	p.scroll.v += p.vramIncrement()
	return value
}

func (p *PPU) writeData(v byte) {
	addr := p.scroll.v & 0x3FFF
	if addr < 0x3F00 {
		p.writeVRAM(addr, v)
	} else {
		p.palette.write(addr, v)
	}
	p.scroll.v += p.vramIncrement()
}

func (p *PPU) readVRAM(addr uint16) byte {
	if addr < 0x2000 {
		return p.cart.ReadCHR(addr)
	}
	table := int((addr / 0x400) % 4)
	return p.nametables[nametableBank(p.mirroring(), table)][addr%0x400]
}

func (p *PPU) writeVRAM(addr uint16, v byte) {
	if addr < 0x2000 {
		p.cart.WriteCHR(addr, v)
		return
	}
	table := int((addr / 0x400) % 4)
	p.nametables[nametableBank(p.mirroring(), table)][addr%0x400] = v
}
