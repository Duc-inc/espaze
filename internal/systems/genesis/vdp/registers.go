package vdp

// Access codes the control port's 6-bit CD field selects (only the
// ones this project implements).
const (
	codeVRAMRead   = 0x00
	codeVRAMWrite  = 0x01
	codeCRAMWrite  = 0x03
	codeVSRAMRead  = 0x04
	codeVSRAMWrite = 0x05
	codeCRAMRead   = 0x08
	codeDMABit     = 0x20
)

// Status register bits (control port read).
const (
	statusVBlank = 1 << 3
)

// writeControl implements the two-word control port protocol: the
// first word carries CD1-CD0 and address bits 13-0, the second CD5-CD2
// and address bits 15-14 - the standard split every Genesis dev
// reference documents this exact way.
func (v *VDP) writeControl(val uint16) {
	if !v.ctrlPending {
		v.ctrlLow = val
		v.ctrlPending = true
		if val&0xC000 == 0x8000 { // a lone register-write word never needs a second word
			v.writeRegisterCommand(val)
			v.ctrlPending = false
		}
		return
	}
	v.ctrlPending = false

	code := byte((val>>2)&0x3C) | byte((v.ctrlLow>>14)&0x03)
	addr := (uint32(val&0x03) << 14) | uint32(v.ctrlLow&0x3FFF)
	v.code = code
	v.addr = uint16(addr)

	if code&codeDMABit != 0 && v.dmaEnabled() {
		v.runDMA()
	}
}

// writeRegisterCommand handles the single-word "set register" form of a
// control port write: bit15=1, bit14=0, bits12-8=register number (0-23),
// bits7-0=value.
func (v *VDP) writeRegisterCommand(val uint16) {
	reg := (val >> 8) & 0x1F
	if reg < 24 {
		v.regs[reg] = byte(val)
	}
}

// ReadStatus implements the control port read.
func (v *VDP) ReadStatus() uint16 {
	return uint16(v.status) | 0x0200 // FIFO-empty bit some games poll for; always report "not busy"
}

func (v *VDP) displayEnabled() bool   { return v.regs[1]&0x40 != 0 }
func (v *VDP) vblankIRQEnabled() bool { return v.regs[1]&0x20 != 0 }
func (v *VDP) dmaEnabled() bool       { return v.regs[1]&0x10 != 0 }
func (v *VDP) h40Mode() bool          { return v.regs[12]&0x01 != 0 }

func (v *VDP) planeABase() uint16 { return uint16(v.regs[2]&0x38) << 10 }
func (v *VDP) planeBBase() uint16 { return uint16(v.regs[4]&0x07) << 13 }
func (v *VDP) spriteTableBase() uint16 {
	return uint16(v.regs[5]&0x7F) << 9
}
func (v *VDP) hScrollTableBase() uint16 { return uint16(v.regs[13]&0x3F) << 10 }
func (v *VDP) autoIncrement() uint16    { return uint16(v.regs[15]) }
func (v *VDP) backdropColor() byte      { return v.regs[7] & 0x3F }

// planeSize returns the current scroll plane's dimensions in tiles -
// always square-ish per axis (32/64/128), independently configurable
// per axis via register 16.
func (v *VDP) planeSize() (width, height int) {
	sizeOf := func(bits byte) int {
		switch bits {
		case 1:
			return 64
		case 3:
			return 128
		default:
			return 32
		}
	}
	width = sizeOf(v.regs[16] & 0x03)
	height = sizeOf((v.regs[16] >> 4) & 0x03)
	return
}

// hScrollMode/vScrollMode decode register 11's scroll granularity bits.
func (v *VDP) hScrollMode() byte      { return v.regs[11] & 0x03 }
func (v *VDP) vScrollPerColumn() bool { return v.regs[11]&0x04 != 0 }
