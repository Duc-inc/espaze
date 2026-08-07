package vdp

// Register 0 (mode control 1) bits.
const (
	reg0VScrollLock = 1 << 7
	reg0HScrollLock = 1 << 6
	reg0MaskColumn0 = 1 << 5
	reg0LineIRQ     = 1 << 4
)

// Register 1 (mode control 2) bits.
const (
	reg1DisplayEnable = 1 << 6
	reg1FrameIRQ      = 1 << 5
	reg1SpriteZoom    = 1 << 1
	reg1SpriteSize16  = 1 << 0
)

// Status register bits (port $BF read).
const (
	statusVBlank         = 1 << 7
	statusSpriteOverflow = 1 << 6
	statusSpriteCollide  = 1 << 5
)

type accessMode int

const (
	accessVRAM accessMode = iota
	accessCRAM
)

// WriteControl implements the two-byte control port ($BF write)
// protocol: the first write latches a low byte, the second combines it
// with the new byte's top 2 "code" bits to mean either a VRAM
// read/write address setup, a CRAM write address setup, or (code 2) a
// register write, where the *first* latched byte is the value and the
// second byte's low nibble is the register number.
func (v *VDP) WriteControl(val byte) {
	if !v.ctrlLatched {
		v.ctrlLow = val
		v.ctrlLatched = true
		return
	}
	v.ctrlLatched = false

	switch val >> 6 {
	case 0: // VRAM read setup
		v.addr = uint16(val&0x3F)<<8 | uint16(v.ctrlLow)
		v.readBuffer = v.vram[v.addr&0x3FFF]
		v.addr = (v.addr + 1) & 0x3FFF
		v.mode = accessVRAM
	case 1: // VRAM write setup
		v.addr = uint16(val&0x3F)<<8 | uint16(v.ctrlLow)
		v.mode = accessVRAM
	case 2: // register write
		reg := val & 0x0F
		if reg <= 10 {
			v.regs[reg] = v.ctrlLow
		}
	default: // CRAM write setup
		v.addr = uint16(val&0x3F)<<8 | uint16(v.ctrlLow)
		v.mode = accessCRAM
	}
}

// ReadStatus implements the status register read (port $BF read),
// which also resets the control port latch and clears every
// event flag - real hardware ties all three to the same read.
func (v *VDP) ReadStatus() byte {
	result := v.status
	v.status = 0
	v.ctrlLatched = false
	v.lineIRQPending = false
	return result
}

// ReadData/WriteData implement the data port (port $BE).
func (v *VDP) ReadData() byte {
	result := v.readBuffer
	v.readBuffer = v.vram[v.addr&0x3FFF]
	v.addr = (v.addr + 1) & 0x3FFF
	v.ctrlLatched = false
	return result
}

func (v *VDP) WriteData(val byte) {
	if v.mode == accessCRAM {
		v.cram.write(byte(v.addr), val)
	} else {
		v.vram[v.addr&0x3FFF] = val
	}
	v.readBuffer = val
	v.addr = (v.addr + 1) & 0x3FFF
	v.ctrlLatched = false
}

func (v *VDP) displayEnabled() bool  { return v.regs[1]&reg1DisplayEnable != 0 }
func (v *VDP) frameIRQEnabled() bool { return v.regs[1]&reg1FrameIRQ != 0 }
func (v *VDP) lineIRQEnabled() bool  { return v.regs[0]&reg0LineIRQ != 0 }
func (v *VDP) spriteHeight() int {
	if v.regs[1]&reg1SpriteSize16 != 0 {
		return 16
	}
	return 8
}

func (v *VDP) nameTableBase() uint16   { return uint16(v.regs[2]&0x0E) << 10 }
func (v *VDP) spriteTableBase() uint16 { return uint16(v.regs[5]&0x7E) << 7 }
func (v *VDP) spritePatternBase() uint16 {
	return uint16(v.regs[6]&0x04) << 11
}
func (v *VDP) backdropColor() byte { return 0x10 | v.regs[7]&0x0F }
func (v *VDP) hScroll() byte       { return v.regs[8] }
func (v *VDP) vScroll() byte       { return v.regs[9] }
