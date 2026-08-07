package memory

// dmaAccess is the subset of the bus a DMA channel copies through -
// defined locally so this file doesn't need to depend on the Bus type
// it's itself part of.
type dmaAccess interface {
	Read8(addr uint32) byte
	Read16(addr uint32) uint16
	Read32(addr uint32) uint32
	Write8(addr uint32, v byte)
	Write16(addr uint32, v uint16)
	Write32(addr uint32, v uint32)
}

// dmaChannel is one of the GBA's 4 DMA channels. Only immediate-start
// transfers actually run in this project - VBlank/HBlank/"special"
// (FIFO-driven) start timing is decoded but never automatically
// retriggered, meaning games that depend on continuous VBlank-synced
// or audio-FIFO-synced DMA won't get it: documented, not silently
// papered over.
type dmaChannel struct {
	src, dst uint32
	count    uint16
	control  uint16
}

func (d *dmaChannel) writeSrcLow(v uint16)  { d.src = d.src&0xFFFF0000 | uint32(v) }
func (d *dmaChannel) writeSrcHigh(v uint16) { d.src = d.src&0x0000FFFF | uint32(v)<<16 }
func (d *dmaChannel) writeDstLow(v uint16)  { d.dst = d.dst&0xFFFF0000 | uint32(v) }
func (d *dmaChannel) writeDstHigh(v uint16) { d.dst = d.dst&0x0000FFFF | uint32(v)<<16 }
func (d *dmaChannel) writeCount(v uint16)   { d.count = v }

func (d *dmaChannel) writeControl(v uint16, bus dmaAccess) {
	wasEnabled := d.control&0x8000 != 0
	d.control = v
	startTiming := (v >> 12) & 0x03
	if v&0x8000 != 0 && !wasEnabled && startTiming == 0 {
		d.run(bus)
	}
}

func (d *dmaChannel) run(bus dmaAccess) {
	word32 := d.control&0x0400 != 0
	srcStep, dstStep := int32(4), int32(4)
	if !word32 {
		srcStep, dstStep = 2, 2
	}
	if (d.control>>7)&0x03 == 1 {
		dstStep = -dstStep
	} else if (d.control>>7)&0x03 == 2 {
		dstStep = 0
	}
	if (d.control>>5)&0x03 == 1 {
		srcStep = -srcStep
	} else if (d.control>>5)&0x03 == 2 {
		srcStep = 0
	}

	count := uint32(d.count)
	if count == 0 {
		count = 0x10000
	}

	src, dst := int64(d.src), int64(d.dst)
	for i := uint32(0); i < count; i++ {
		if word32 {
			bus.Write32(uint32(dst), bus.Read32(uint32(src)))
		} else {
			bus.Write16(uint32(dst), bus.Read16(uint32(src)))
		}
		src += int64(srcStep)
		dst += int64(dstStep)
	}

	if d.control&0x0200 == 0 { // not repeating: auto-disable
		d.control &^= 0x8000
	}
}
