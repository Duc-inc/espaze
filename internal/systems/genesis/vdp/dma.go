package vdp

// MemoryReader is the 68000-visible memory space DMA reads from for a
// "68k to VDP" transfer - real hardware also stalls the CPU while this
// runs, which this project's instantaneous DMA (like every other DMA
// in this codebase) doesn't reproduce.
type MemoryReader interface {
	Read16(addr uint32) uint16
}

func (v *VDP) dmaLength() uint32 {
	length := uint32(v.regs[19]) | uint32(v.regs[20])<<8
	if length == 0 {
		length = 0x10000
	}
	return length
}

func (v *VDP) dmaSource() uint32 {
	return uint32(v.regs[21])<<1 | uint32(v.regs[22])<<9 | uint32(v.regs[23]&0x7F)<<17
}

// runDMA dispatches on register 23's top 2 bits: 68k-to-VDP copy (bit7
// clear), VRAM fill (bit7 set, bit6 clear, armed here but actually
// performed on the *next* data-port write - see WriteData's caller in
// io.go, which checks dmaFillArmed), or VRAM-to-VRAM copy (both set).
func (v *VDP) runDMA() {
	switch {
	case v.regs[23]&0x80 == 0:
		v.dmaMemoryToVDP()
	case v.regs[23]&0x40 == 0:
		v.dmaFillArmed = true
	default:
		v.dmaVRAMCopy()
	}
}

func (v *VDP) dmaMemoryToVDP() {
	if v.mem == nil {
		return
	}
	length := v.dmaLength()
	src := v.dmaSource()
	inc := v.autoIncrement()

	for i := uint32(0); i < length; i++ {
		word := v.mem.Read16(src)
		v.writeDataRaw(word)
		src += 2
		_ = inc
	}
	v.regs[19], v.regs[20] = 0, 0
}

// dmaFill is called from WriteData when a fill DMA is armed: val's
// upper byte becomes the fill byte, repeated for the whole run.
func (v *VDP) dmaFill(val uint16) {
	length := v.dmaLength()
	inc := v.autoIncrement()
	fillByte := byte(val >> 8)

	v.vram[v.addr&0xFFFF] = fillByte
	for i := uint32(1); i < length; i++ {
		v.addr += inc
		v.vram[v.addr&0xFFFF] = fillByte
	}
	v.dmaFillArmed = false
	v.regs[19], v.regs[20] = 0, 0
}

func (v *VDP) dmaVRAMCopy() {
	length := v.dmaLength()
	src := uint16(v.dmaSource())
	inc := v.autoIncrement()

	for i := uint32(0); i < length; i++ {
		v.vram[v.addr&0xFFFF] = v.vram[src&0xFFFF]
		src++
		v.addr += inc
	}
	v.regs[19], v.regs[20] = 0, 0
}

// writeDataRaw is WriteData's body without the DMA-arming check, used
// by the memory-to-VDP transfer to write each word it copies.
func (v *VDP) writeDataRaw(val uint16) {
	switch v.code &^ codeDMABit {
	case codeCRAMWrite:
		v.palette.write(byte(v.addr>>1), val)
	case codeVSRAMWrite:
		idx := (v.addr >> 1) & 0x3F
		if idx < 40 {
			v.vsram[idx] = val
		}
	default:
		v.vram[v.addr&0xFFFE] = byte(val >> 8)
		v.vram[(v.addr+1)&0xFFFF] = byte(val)
	}
	v.addr += v.autoIncrement()
}
