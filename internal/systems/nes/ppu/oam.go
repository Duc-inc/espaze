package ppu

// oam is the PPU's sprite attribute memory: 64 sprites, 4 bytes each
// (Y, tile index, attributes, X), reachable from the CPU via
// OAMADDR/OAMDATA or in bulk through $4014 OAM DMA.
type oam struct {
	data [256]byte
}

func (o *oam) readByte(addr byte) byte     { return o.data[addr] }
func (o *oam) writeByte(addr byte, v byte) { o.data[addr] = v }
