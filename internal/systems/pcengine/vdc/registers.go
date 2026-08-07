package vdc

// Register indices this project implements.
const (
	regMAWR = 0x00
	regMARR = 0x01
	regVRAM = 0x02 // data port: write uses MAWR, read uses MARR
	regCR   = 0x05
	regRCR  = 0x06
	regBXR  = 0x07
	regBYR  = 0x08
	regSATB = 0x13
)

// SelectRegister implements the control port write that picks which
// register the following data-port writes/reads target.
func (v *VDC) SelectRegister(reg byte) {
	v.selectedReg = reg & 0x1F
	v.writeHiNext = false
}

// WriteDataLow/WriteDataHigh implement the two-byte data port.
func (v *VDC) WriteDataLow(b byte) {
	if int(v.selectedReg) >= len(v.regs) {
		return
	}
	if v.selectedReg == regVRAM {
		v.writeVRAMLow(b)
		return
	}
	v.regs[v.selectedReg] = v.regs[v.selectedReg]&0xFF00 | uint16(b)
}

func (v *VDC) WriteDataHigh(b byte) {
	if int(v.selectedReg) >= len(v.regs) {
		return
	}
	if v.selectedReg == regVRAM {
		v.writeVRAMHigh(b)
		return
	}
	v.regs[v.selectedReg] = v.regs[v.selectedReg]&0x00FF | uint16(b)<<8
	if v.selectedReg == regSATB {
		v.runSATBDMA()
	}
}

func (v *VDC) writeVRAMLow(b byte) { v.vramLowLatch = b }

func (v *VDC) writeVRAMHigh(b byte) {
	word := uint16(v.vramLowLatch) | uint16(b)<<8
	v.vram[v.regs[regMAWR]&0x7FFF] = word
	v.regs[regMAWR]++
}

// ReadDataLow/ReadDataHigh implement the data port's read direction -
// simplified to read immediately from MARR rather than reproducing
// real hardware's one-word read-ahead buffer.
func (v *VDC) ReadDataLow() byte {
	return byte(v.vram[v.regs[regMARR]&0x7FFF])
}

func (v *VDC) ReadDataHigh() byte {
	word := v.vram[v.regs[regMARR]&0x7FFF]
	v.regs[regMARR]++
	return byte(word >> 8)
}

// runSATBDMA copies 256 words from VRAM (starting at the address just
// latched into register $13) into the internal sprite attribute
// table - real hardware performs this at the next VBlank rather than
// instantly; every DMA in this project simplifies that away the same
// way.
func (v *VDC) runSATBDMA() {
	base := v.regs[regSATB]
	for i := 0; i < len(v.sat); i++ {
		v.sat[i] = v.vram[(base+uint16(i))&0x7FFF]
	}
}
