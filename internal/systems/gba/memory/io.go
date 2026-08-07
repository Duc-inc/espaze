package memory

// writeIO8 handles the few I/O registers real games commonly touch
// with 8-bit writes - the FIFO ports. Every other register in this
// project's implementation expects the standard 16/32-bit access GBA
// software is documented to use, so byte writes elsewhere are dropped
// rather than approximated.
func (b *Bus) writeIO8(addr uint32, v byte) {
	switch addr & 0xFFFF {
	case 0x0A0, 0x0A1, 0x0A2, 0x0A3:
		b.sound.WriteFIFOA(v)
	case 0x0A4, 0x0A5, 0x0A6, 0x0A7:
		b.sound.WriteFIFOB(v)
	}
}

func (b *Bus) writeIO(addr uint32, v uint16) {
	switch addr & 0xFFFF {
	case 0x000:
		b.video.WriteDISPCNT(v)
	case 0x004:
		b.video.WriteDISPSTAT(v)
	case 0x008:
		b.video.WriteBGCNT(0, v)
	case 0x00A:
		b.video.WriteBGCNT(1, v)
	case 0x00C:
		b.video.WriteBGCNT(2, v)
	case 0x00E:
		b.video.WriteBGCNT(3, v)
	case 0x010:
		b.video.WriteBGHOFS(0, v)
	case 0x012:
		b.video.WriteBGVOFS(0, v)
	case 0x014:
		b.video.WriteBGHOFS(1, v)
	case 0x016:
		b.video.WriteBGVOFS(1, v)
	case 0x018:
		b.video.WriteBGHOFS(2, v)
	case 0x01A:
		b.video.WriteBGVOFS(2, v)
	case 0x01C:
		b.video.WriteBGHOFS(3, v)
	case 0x01E:
		b.video.WriteBGVOFS(3, v)
	case 0x082:
		b.sound.WriteSoundCntH(v)
	case 0x0A0, 0x0A2:
		b.sound.WriteFIFOA(byte(v))
	case 0x0A4, 0x0A6:
		b.sound.WriteFIFOB(byte(v))
	case 0x100:
		b.tm.writeReload(0, v)
	case 0x102:
		b.tm.writeControl(0, v)
	case 0x104:
		b.tm.writeReload(1, v)
	case 0x106:
		b.tm.writeControl(1, v)
	case 0x108:
		b.tm.writeReload(2, v)
	case 0x10A:
		b.tm.writeControl(2, v)
	case 0x10C:
		b.tm.writeReload(3, v)
	case 0x10E:
		b.tm.writeControl(3, v)
	case 0x200:
		b.irq.writeIE(v)
	case 0x202:
		b.irq.acknowledgeIF(v)
	case 0x208:
		b.irq.writeIME(v)
	default:
		b.writeDMARegister(addr&0xFFFF, v)
	}
}

func (b *Bus) writeDMARegister(offset uint32, v uint16) {
	if offset < 0x0B0 || offset > 0x0DF {
		return
	}
	ch := int(offset-0x0B0) / 12
	if ch > 3 {
		return
	}
	switch (offset - 0x0B0) % 12 {
	case 0:
		b.dma[ch].writeSrcLow(v)
	case 2:
		b.dma[ch].writeSrcHigh(v)
	case 4:
		b.dma[ch].writeDstLow(v)
	case 6:
		b.dma[ch].writeDstHigh(v)
	case 8:
		b.dma[ch].writeCount(v)
	case 10:
		b.dma[ch].writeControl(v, b)
	}
}

func (b *Bus) readIO(addr uint32) uint16 {
	switch addr & 0xFFFF {
	case 0x000:
		return b.video.ReadDISPCNT()
	case 0x100:
		return b.tm.readCounter(0)
	case 0x104:
		return b.tm.readCounter(1)
	case 0x108:
		return b.tm.readCounter(2)
	case 0x10C:
		return b.tm.readCounter(3)
	case 0x130:
		return b.kp.read()
	case 0x200:
		return b.irq.ie
	case 0x202:
		return b.irq.readIF()
	case 0x208:
		if b.irq.ime {
			return 1
		}
		return 0
	default:
		return 0
	}
}
