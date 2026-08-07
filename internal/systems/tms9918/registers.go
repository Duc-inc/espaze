package tms9918

// WriteControl implements the real TMS9918 control port protocol: two
// consecutive writes. The first latches a byte; the second either
// writes a register (bit7 set: register = bits0-2, value = the
// latched byte) or completes a 14-bit VRAM address (bit7 clear: bits
// 0-5 are the address's high bits, bit6 selects read/write mode) -
// this exact two-byte latch is real, well-documented hardware
// behavior, unlike this project's fixed VRAM table layout.
func (t *TMS9918) WriteControl(v byte) {
	if !t.latched {
		t.addrLatch = v
		t.latched = true
		return
	}
	t.latched = false

	if v&0x80 != 0 {
		reg := v & 0x07
		if reg < 8 {
			t.regs[reg] = t.addrLatch
		}
		return
	}

	t.addr = uint16(v&0x3F)<<8 | uint16(t.addrLatch)
	t.writeMode = v&0x40 != 0
	if !t.writeMode {
		t.readBuffer = t.vram[t.addr&0x3FFF]
		t.addr++
	}
}

// WriteData implements the data port write.
func (t *TMS9918) WriteData(v byte) {
	t.vram[t.addr&0x3FFF] = v
	t.addr++
}

// ReadData implements the data port read - real hardware buffers one
// byte ahead, refilling the buffer on every read.
func (t *TMS9918) ReadData() byte {
	v := t.readBuffer
	t.readBuffer = t.vram[t.addr&0x3FFF]
	t.addr++
	return v
}

// ReadStatus implements the status port read: bit7 is the VBlank flag,
// cleared by this read - real hardware's own behavior.
func (t *TMS9918) ReadStatus() byte {
	v := t.status
	t.status &^= 0x80
	t.latched = false
	return v
}
