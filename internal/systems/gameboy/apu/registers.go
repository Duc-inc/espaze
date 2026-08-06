package apu

// readMasks are the bits that always read back as 1 regardless of what
// was written, per the DMG's documented sound register behavior. Keyed
// by offset from 0xFF10.
var readMasks = [23]byte{
	0x80, 0x3F, 0x00, 0xFF, 0xBF, // NR10-NR14 (0xFF14 has no NR15 gap here)
	0xFF, 0x3F, 0x00, 0xFF, 0xBF, // NR20(unused)-NR24
	0x7F, 0xFF, 0x9F, 0xFF, 0xBF, // NR30-NR34
	0xFF, 0xFF, 0x00, 0x00, 0xBF, // NR40(unused)-NR44
	0x00, 0x00, 0x70, // NR50-NR52
}

// WriteRegister implements the CPU writing 0xFF10-0xFF26 (control) and
// 0xFF30-0xFF3F (wave RAM).
func (a *APU) WriteRegister(addr uint16, v byte) {
	if addr >= 0xFF30 && addr <= 0xFF3F {
		a.ch3.ram[addr-0xFF30] = v
		return
	}
	if !a.powerOn && addr != 0xFF26 {
		return
	}

	switch addr {
	case 0xFF10:
		a.ch1.sweep.writeRegister(v)
	case 0xFF11:
		a.ch1.duty = v >> 6
		a.ch1.length.setFromRegister(int(v & 0x3F))
	case 0xFF12:
		a.ch1.envelope.writeRegister(v)
		if !a.ch1.envelope.dacEnabled() {
			a.ch1.enabled = false
		}
	case 0xFF13:
		a.ch1.frequency = (a.ch1.frequency & 0x700) | uint16(v)
	case 0xFF14:
		a.ch1.frequency = (a.ch1.frequency & 0xFF) | (uint16(v&0x07) << 8)
		a.ch1.length.enabled = v&0x40 != 0
		if v&0x80 != 0 {
			a.ch1.trigger()
		}

	case 0xFF16:
		a.ch2.duty = v >> 6
		a.ch2.length.setFromRegister(int(v & 0x3F))
	case 0xFF17:
		a.ch2.envelope.writeRegister(v)
		if !a.ch2.envelope.dacEnabled() {
			a.ch2.enabled = false
		}
	case 0xFF18:
		a.ch2.frequency = (a.ch2.frequency & 0x700) | uint16(v)
	case 0xFF19:
		a.ch2.frequency = (a.ch2.frequency & 0xFF) | (uint16(v&0x07) << 8)
		a.ch2.length.enabled = v&0x40 != 0
		if v&0x80 != 0 {
			a.ch2.trigger()
		}

	case 0xFF1A:
		a.ch3.dacEnabled = v&0x80 != 0
		if !a.ch3.dacEnabled {
			a.ch3.enabled = false
		}
	case 0xFF1B:
		a.ch3.length.setFromRegister(int(v))
	case 0xFF1C:
		a.ch3.volumeCode = (v >> 5) & 0x03
	case 0xFF1D:
		a.ch3.frequency = (a.ch3.frequency & 0x700) | uint16(v)
	case 0xFF1E:
		a.ch3.frequency = (a.ch3.frequency & 0xFF) | (uint16(v&0x07) << 8)
		a.ch3.length.enabled = v&0x40 != 0
		if v&0x80 != 0 {
			a.ch3.trigger()
		}

	case 0xFF20:
		a.ch4.length.setFromRegister(int(v & 0x3F))
	case 0xFF21:
		a.ch4.envelope.writeRegister(v)
		if !a.ch4.envelope.dacEnabled() {
			a.ch4.enabled = false
		}
	case 0xFF22:
		a.ch4.shiftClockFreq = v >> 4
		a.ch4.widthMode = v&0x08 != 0
		a.ch4.divisorCode = v & 0x07
	case 0xFF23:
		a.ch4.length.enabled = v&0x40 != 0
		if v&0x80 != 0 {
			a.ch4.trigger()
		}

	case 0xFF24:
		a.masterLeft = (v >> 4) & 0x07
		a.masterRight = v & 0x07
	case 0xFF25:
		a.panning = v
	case 0xFF26:
		wasOn := a.powerOn
		a.powerOn = v&0x80 != 0
		if wasOn && !a.powerOn {
			a.Reset()
		}
	}
}

// ReadRegister implements the CPU reading 0xFF10-0xFF26 and wave RAM.
func (a *APU) ReadRegister(addr uint16) byte {
	if addr >= 0xFF30 && addr <= 0xFF3F {
		return a.ch3.ram[addr-0xFF30]
	}
	if addr == 0xFF26 {
		return a.statusByte()
	}
	if addr < 0xFF10 || addr > 0xFF25 {
		return 0xFF
	}
	return readMasks[addr-0xFF10]
}

func (a *APU) statusByte() byte {
	status := readMasks[0xFF26-0xFF10]
	if a.powerOn {
		status |= 0x80
	}
	if a.ch1.enabled {
		status |= 0x01
	}
	if a.ch2.enabled {
		status |= 0x02
	}
	if a.ch3.enabled {
		status |= 0x04
	}
	if a.ch4.enabled {
		status |= 0x08
	}
	return status
}
