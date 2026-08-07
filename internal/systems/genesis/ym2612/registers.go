package ym2612

// WriteAddress1/WriteAddress2 latch which register the next
// WriteData1/WriteData2 call targets - real hardware's two independent
// address/data port pairs, one covering channels 1-3 (and the global
// registers), the other channels 4-6.
func (y *YM2612) WriteAddress1(v byte) { y.addr1 = v }
func (y *YM2612) WriteAddress2(v byte) { y.addr2 = v }

func (y *YM2612) WriteData1(v byte) { y.writeRegister(y.addr1, v, 0) }
func (y *YM2612) WriteData2(v byte) { y.writeRegister(y.addr2, v, 3) }

// writeRegister dispatches a register write to the right channel/
// operator. chanBase is 0 for part 1 (channels 0-2) or 3 for part 2
// (channels 3-5, i.e. the chip's "channel 4-6").
func (y *YM2612) writeRegister(addr, v byte, chanBase int) {
	if addr == 0x28 { // key on/off is global, only ever written via part 1
		y.writeKeyOnOff(v)
		return
	}

	switch {
	case addr >= 0x30 && addr < 0xA0:
		y.writeOperatorRegister(addr, v, chanBase)
	case addr >= 0xA0 && addr <= 0xA2:
		y.channels[chanBase+int(addr&0x03)].writeFNumLow(v)
	case addr >= 0xA4 && addr <= 0xA6:
		y.channels[chanBase+int(addr&0x03)].writeFNumHighBlock(v)
	case addr >= 0xB0 && addr <= 0xB2:
		y.channels[chanBase+int(addr&0x03)].writeFeedbackAlgorithm(v)
	case addr >= 0xB4 && addr <= 0xB6:
		y.channels[chanBase+int(addr&0x03)].writePan(v)
	}
}

func (y *YM2612) writeOperatorRegister(addr, v byte, chanBase int) {
	offset := addr & 0x0F
	ch := int(offset & 0x03)
	if ch == 3 {
		return // unused slot
	}
	op := (offset >> 2) & 0x03
	o := &y.channels[chanBase+ch].ops[op]

	switch addr & 0xF0 {
	case 0x30:
		o.writeMul(v)
	case 0x40:
		o.writeTL(v)
	case 0x50:
		o.writeAR(v)
	case 0x60:
		o.writeD1R(v)
	case 0x70:
		o.writeD2R(v)
	case 0x80:
		o.writeSLRR(v)
	}
}

// writeKeyOnOff implements $28: bits0-1 select the channel within its
// part, bit2 selects part 2 (channels 3-5) over part 1, bits4-7 gate
// each of the 4 operators on or off.
func (y *YM2612) writeKeyOnOff(v byte) {
	chanInPart := int(v & 0x03)
	if chanInPart == 3 {
		return
	}
	ch := chanInPart
	if v&0x04 != 0 {
		ch += 3
	}

	for op := byte(0); op < 4; op++ {
		on := v&(0x10<<op) != 0
		if on {
			y.channels[ch].keyOnOp(op)
		} else {
			y.channels[ch].keyOffOp(op)
		}
	}
}
