package apu

// WriteRegister implements CPU access to $4000-$4013 and $4015/$4017.
// $4014 (OAM DMA) isn't APU state and is handled at the bus level.
func (a *APU) WriteRegister(addr uint16, v byte) {
	switch addr {
	case 0x4000:
		a.pulse1.writeControl(v)
	case 0x4001:
		a.pulse1.sweep.write(v)
	case 0x4002:
		a.pulse1.writeTimerLow(v)
	case 0x4003:
		a.pulse1.writeTimerHighAndLength(v)
	case 0x4004:
		a.pulse2.writeControl(v)
	case 0x4005:
		a.pulse2.sweep.write(v)
	case 0x4006:
		a.pulse2.writeTimerLow(v)
	case 0x4007:
		a.pulse2.writeTimerHighAndLength(v)
	case 0x4008:
		a.triangle.writeLinear(v)
	case 0x400A:
		a.triangle.writeTimerLow(v)
	case 0x400B:
		a.triangle.writeTimerHighAndLength(v)
	case 0x400C:
		a.noise.writeControl(v)
	case 0x400E:
		a.noise.writeMode(v)
	case 0x400F:
		a.noise.writeLength(v)
	case 0x4010:
		a.dmc.writeControl(v)
	case 0x4011:
		a.dmc.writeDirectLoad(v)
	case 0x4012:
		a.dmc.writeSampleAddr(v)
	case 0x4013:
		a.dmc.writeSampleLength(v)
	case 0x4015:
		a.writeStatus(v)
	case 0x4017:
		a.seq.write(v, a)
	}
}

// writeStatus handles $4015: enabling/disabling each channel (disabling
// force-silences it by zeroing its length counter) and acknowledging
// any pending DMC IRQ.
func (a *APU) writeStatus(v byte) {
	a.pulse1.setEnabled(v&0x01 != 0)
	a.pulse2.setEnabled(v&0x02 != 0)
	a.triangle.setEnabled(v&0x04 != 0)
	a.noise.setEnabled(v&0x08 != 0)
	a.dmc.setEnabled(v&0x10 != 0)
	a.dmc.irqFlag = false
}

// ReadRegister implements CPU access to $4015 (the only readable APU
// register - everything else reads back open bus, which the caller is
// expected to supply itself).
func (a *APU) ReadRegister(addr uint16) byte {
	if addr != 0x4015 {
		return 0
	}
	var v byte
	if a.pulse1.length.active() {
		v |= 0x01
	}
	if a.pulse2.length.active() {
		v |= 0x02
	}
	if a.triangle.length.active() {
		v |= 0x04
	}
	if a.noise.length.active() {
		v |= 0x08
	}
	if a.dmc.active() {
		v |= 0x10
	}
	if a.seq.irqFlag {
		v |= 0x40
	}
	if a.dmc.irqFlag {
		v |= 0x80
	}
	a.seq.irqFlag = false
	return v
}
