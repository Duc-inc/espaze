// Package di implements the GameCube's DVD Interface: the real
// memory-mapped registers a game uses to read the disc, replacing
// this project's other "just call disc.ParseHeader directly" approach
// with the actual hardware path real games use. Addresses, the DMA
// register layout, the 0xA8 read-sector command, and DISR's bit table
// come from a public hardware register reference (YAGCD chapter 5,
// section 5.7, "DI - DVD Interface"), looked up directly against that
// source.
package di

const (
	Base = 0xCC006000
	Size = 0x40

	regDISR     = 0x00
	regCMDBUF0  = 0x08
	regCMDBUF1  = 0x0C
	regCMDBUF2  = 0x10
	regDIMAR    = 0x14
	regDILENGTH = 0x18
	regDICR     = 0x1C

	cmdReadSector = 0xA8

	// DISR bits (YAGCD 5.7). BRKINT/TCINT/DEINT are status bits,
	// write-1-to-clear; their paired MASK bit gates whether that status
	// reaches PI's DI interrupt line (see Interrupting). BRK requests a
	// break of the current transfer - not modeled, this project's DMA
	// always completes in one execute() call.
	bitBRKINT     = 1 << 6
	bitBRKINTMASK = 1 << 5
	bitTCINT      = 1 << 4
	bitTCINTMASK  = 1 << 3
	bitDEINT      = 1 << 2
	bitDEINTMASK  = 1 << 1
	bitBRK        = 1 << 0
)

// MemWriter is the subset of main memory access DI needs to DMA disc
// data into.
type MemWriter interface {
	Write8(addr uint32, v byte)
}

// DI holds the DVD Interface's registers plus the disc image it reads
// from - this project's stand-in for a real disc drive.
type DI struct {
	image []byte
	mem   MemWriter

	cmdBuf0, cmdBuf1, cmdBuf2 uint32
	dimar, dilength           uint32

	// tcint/deint/brkint are DISR's status bits; this project only ever
	// sets tcint (execute() always "succeeds" - no error or break path
	// is modeled), but all three plus their mask bits are still stored
	// and readable/writable for a game that polls DISR's full layout.
	tcint, tcintMask   bool
	deint, deintMask   bool
	brkint, brkintMask bool
}

// New returns a DI serving reads from image, DMA'ing into mem.
func New(image []byte, mem MemWriter) *DI {
	return &DI{image: image, mem: mem}
}

// Interrupting reports whether any of DI's status bits is both active
// and unmasked - the level-triggered cause signal pi.PI's DI bit
// reports (see gamecube.go's Step). Real hardware ORs all three
// status/mask pairs onto the same interrupt line.
func (d *DI) Interrupting() bool {
	return (d.tcint && d.tcintMask) || (d.deint && d.deintMask) || (d.brkint && d.brkintMask)
}

func (d *DI) Read32(offset uint32) uint32 {
	switch offset {
	case regDISR:
		var v uint32
		if d.brkint {
			v |= bitBRKINT
		}
		if d.brkintMask {
			v |= bitBRKINTMASK
		}
		if d.tcint {
			v |= bitTCINT
		}
		if d.tcintMask {
			v |= bitTCINTMASK
		}
		if d.deint {
			v |= bitDEINT
		}
		if d.deintMask {
			v |= bitDEINTMASK
		}
		return v
	case regCMDBUF0:
		return d.cmdBuf0
	case regCMDBUF1:
		return d.cmdBuf1
	case regCMDBUF2:
		return d.cmdBuf2
	case regDIMAR:
		return d.dimar
	case regDILENGTH:
		return d.dilength
	default:
		return 0
	}
}

func (d *DI) Write32(offset uint32, val uint32) {
	switch offset {
	case regDISR:
		if val&bitBRKINT != 0 {
			d.brkint = false // write-1-to-clear
		}
		if val&bitTCINT != 0 {
			d.tcint = false // write-1-to-clear
		}
		if val&bitDEINT != 0 {
			d.deint = false // write-1-to-clear
		}
		d.brkintMask = val&bitBRKINTMASK != 0
		d.tcintMask = val&bitTCINTMASK != 0
		d.deintMask = val&bitDEINTMASK != 0
	case regCMDBUF0:
		d.cmdBuf0 = val
	case regCMDBUF1:
		d.cmdBuf1 = val
	case regCMDBUF2:
		d.cmdBuf2 = val
	case regDIMAR:
		d.dimar = val
	case regDILENGTH:
		d.dilength = val
	case regDICR:
		if val&1 != 0 { // TSTART
			d.execute()
		}
	}
}

// execute runs whatever command is currently in cmdBuf0's top byte.
// Only 0xA8 (read sector) is implemented - real hardware's other
// commands (inquiry, seek, audio streaming) aren't modeled.
func (d *DI) execute() {
	switch byte(d.cmdBuf0 >> 24) {
	case cmdReadSector:
		srcOffset := d.cmdBuf1 << 2 // DICMDBUF1 is a 32-bit-word offset
		length := d.cmdBuf2
		for i := uint32(0); i < length; i++ {
			var b byte
			if int(srcOffset+i) < len(d.image) {
				b = d.image[srcOffset+i]
			}
			d.mem.Write8(d.dimar+i, b)
		}
	}
	d.tcint = true
}
