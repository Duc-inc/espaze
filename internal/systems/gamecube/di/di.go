// Package di implements the GameCube's DVD Interface: the real
// memory-mapped registers a game uses to read the disc, replacing
// this project's other "just call disc.ParseHeader directly" approach
// with the actual hardware path real games use. Addresses, the DMA
// register layout, and the 0xA8 read-sector command come from a
// public hardware register reference (YAGCD chapter 5, "DI - DVD
// Interface"); DISR's exact status/mask bit numbering wasn't cleanly
// legible in that source, so it's modeled on the well-known status-
// bit-then-mask-bit pairing convention this kind of register commonly
// uses, not a directly confirmed bit table.
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

	bitTCINT = 1 << 0 // transfer-complete interrupt status
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
	tcint                     bool
}

// New returns a DI serving reads from image, DMA'ing into mem.
func New(image []byte, mem MemWriter) *DI {
	return &DI{image: image, mem: mem}
}

func (d *DI) Read32(offset uint32) uint32 {
	switch offset {
	case regDISR:
		var v uint32
		if d.tcint {
			v |= bitTCINT
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
		if val&bitTCINT != 0 {
			d.tcint = false // write-1-to-clear
		}
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
