package vdp

// WriteControl implements the control port write.
func (v *VDP) WriteControl(val uint16) { v.writeControl(val) }

// WriteData implements the data port write: routes to VRAM, CRAM or
// VSRAM depending on the code the last control port command set up,
// then advances the address by the auto-increment register. If a VRAM
// fill DMA is currently armed (see dma.go), this write's upper byte is
// the fill value instead of an ordinary write - that's genuinely how
// real hardware triggers a fill, not a simplification.
func (v *VDP) WriteData(val uint16) {
	if v.dmaFillArmed {
		v.dmaFill(val)
		return
	}
	v.writeDataRaw(val)
}

// ReadData implements the data port read.
func (v *VDP) ReadData() uint16 {
	var result uint16
	switch v.code {
	case codeCRAMRead:
		result = v.palette.read(byte(v.addr >> 1))
	case codeVSRAMRead:
		idx := (v.addr >> 1) & 0x3F
		if idx < 40 {
			result = v.vsram[idx]
		}
	default:
		result = uint16(v.vram[v.addr&0xFFFE])<<8 | uint16(v.vram[(v.addr+1)&0xFFFF])
	}
	v.addr += v.autoIncrement()
	return result
}

// vramWord/vramByte are internal helpers the renderer uses to read
// pattern/name-table data directly, bypassing the code/address port
// state (which only matters for CPU-issued reads/writes).
func (v *VDP) vramWord(addr uint16) uint16 {
	return uint16(v.vram[addr&0xFFFE])<<8 | uint16(v.vram[(addr+1)&0xFFFF])
}
