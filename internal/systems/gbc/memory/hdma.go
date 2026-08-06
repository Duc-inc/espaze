package memory

// hdma implements $FF51-$FF55, the CGB's fast VRAM DMA controller. Real
// hardware can spread an "H-Blank DMA" transfer across many H-Blank
// periods for smooth mid-frame effects (split-screen gradients and
// similar); this implementation simplifies both HDMA and general-purpose
// DMA to one immediate transfer, which preserves the *result* - every
// byte ends up in the right place - without the exact per-scanline
// timing a handful of games use for those transition effects.
type hdma struct {
	srcHi, srcLo byte
	dstHi, dstLo byte
}

func (h *hdma) writeSrcHi(v byte) { h.srcHi = v }
func (h *hdma) writeSrcLo(v byte) { h.srcLo = v & 0xF0 }
func (h *hdma) writeDstHi(v byte) { h.dstHi = v & 0x1F }
func (h *hdma) writeDstLo(v byte) { h.dstLo = v & 0xF0 }

func (h *hdma) source() uint16 { return uint16(h.srcHi)<<8 | uint16(h.srcLo) }
func (h *hdma) dest() uint16   { return 0x8000 | uint16(h.dstHi)<<8 | uint16(h.dstLo) }

// transfer performs the (length+1)*0x10-byte copy a $FF55 write
// requests, reading through read (the full CPU bus, so any memory
// region can be a source) and writing through writeVRAM (the PPU's
// currently-selected VRAM bank).
func (h *hdma) transfer(length byte, read func(uint16) byte, writeVRAM func(uint16, byte)) {
	count := (int(length&0x7F) + 1) * 0x10
	src, dst := h.source(), h.dest()
	for i := 0; i < count; i++ {
		writeVRAM(dst+uint16(i), read(src+uint16(i)))
	}
}
