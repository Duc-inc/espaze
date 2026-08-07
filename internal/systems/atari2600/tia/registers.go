package tia

// currentX returns the visible-pixel x-coordinate the beam is
// currently at, clamped to 0 during horizontal blank - used by the
// RESxx strobes, which position an object wherever the beam happens to
// be at the moment they're written (real hardware adds a small further
// delay this simplifies away).
func (t *TIA) currentX() int {
	if t.clock < hblankClocks {
		return 0
	}
	return t.clock - hblankClocks
}

// WriteRegister implements a CPU write to the TIA's address space
// ($00-$2C, mirrored - the caller is expected to have already masked
// the address to that range).
func (t *TIA) WriteRegister(addr byte, v byte) {
	switch addr {
	case 0x00:
		t.vsync = v&0x02 != 0
	case 0x01:
		t.vblank = v&0x02 != 0
	case 0x02:
		t.wsync = true
	case 0x04:
		t.m0.writeSize(v >> 4)
	case 0x05:
		t.m1.writeSize(v >> 4)
	case 0x06:
		t.p0.color = v
	case 0x07:
		t.p1.color = v
	case 0x08:
		t.colupf = v
	case 0x09:
		t.bg = v
	case 0x0A:
		t.pf.writeCTRLPF(v)
		t.bl.writeSize(v >> 4)
	case 0x0B:
		t.p0.reflect = v&0x08 != 0
	case 0x0C:
		t.p1.reflect = v&0x08 != 0
	case 0x0D:
		t.pf.pf0 = v
	case 0x0E:
		t.pf.pf1 = v
	case 0x0F:
		t.pf.pf2 = v
	case 0x10:
		t.p0.reset(t.currentX())
	case 0x11:
		t.p1.reset(t.currentX())
	case 0x12:
		t.m0.reset(t.currentX())
	case 0x13:
		t.m1.reset(t.currentX())
	case 0x14:
		t.bl.reset(t.currentX())
	case 0x15:
		t.a0.writeAUDC(v)
	case 0x16:
		t.a1.writeAUDC(v)
	case 0x17:
		t.a0.writeAUDF(v)
	case 0x18:
		t.a1.writeAUDF(v)
	case 0x19:
		t.a0.writeAUDV(v)
	case 0x1A:
		t.a1.writeAUDV(v)
	case 0x1B:
		t.p0.grp = v
	case 0x1C:
		t.p1.grp = v
	case 0x1D:
		t.m0.enabled = v&0x02 != 0
	case 0x1E:
		t.m1.enabled = v&0x02 != 0
	case 0x1F:
		t.bl.enabled = v&0x02 != 0
	case 0x20:
		t.hmp0 = v
	case 0x21:
		t.hmp1 = v
	case 0x22:
		t.hmm0 = v
	case 0x23:
		t.hmm1 = v
	case 0x24:
		t.hmbl = v
	case 0x2A:
		t.p0.applyMotion(t.hmp0)
		t.p1.applyMotion(t.hmp1)
		t.m0.applyMotion(t.hmm0)
		t.m1.applyMotion(t.hmm1)
		t.bl.applyMotion(t.hmbl)
	case 0x2B:
		t.hmp0, t.hmp1, t.hmm0, t.hmm1, t.hmbl = 0, 0, 0, 0, 0
	case 0x2C:
		t.cxm0p, t.cxm1p, t.cxp0fb, t.cxp1fb = 0, 0, 0, 0
		t.cxm0fb, t.cxm1fb, t.cxblpf, t.cxppmm = 0, 0, 0, 0
	}
}

// ReadRegister implements a CPU read from the TIA's address space
// ($00-$0D, mirrored).
func (t *TIA) ReadRegister(addr byte) byte {
	switch addr {
	case 0x00:
		return t.cxm0p
	case 0x01:
		return t.cxm1p
	case 0x02:
		return t.cxp0fb
	case 0x03:
		return t.cxp1fb
	case 0x04:
		return t.cxm0fb
	case 0x05:
		return t.cxm1fb
	case 0x06:
		return t.cxblpf
	case 0x07:
		return t.cxppmm
	case 0x08, 0x09, 0x0A, 0x0B:
		return 0 // paddle pots: not implemented, digital joysticks only
	case 0x0C, 0x0D:
		return t.inputLatches[addr-0x08]
	default:
		return 0
	}
}

// SetButton feeds the joystick fire button's state into INPT4/INPT5
// (player 0/1), active-low in bit7 exactly like the real port.
func (t *TIA) SetButton(player int, pressed bool) {
	idx := 4
	if player == 1 {
		idx = 5
	}
	if pressed {
		t.inputLatches[idx] = 0x00
	} else {
		t.inputLatches[idx] = 0x80
	}
}
