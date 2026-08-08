// Package exi implements the GameCube's External Interface: the real
// memory-mapped registers that drive memory cards and other EXI-bus
// devices. Base address, channel count/stride, and the CSR/EXIDATA
// register layout (TSTART, CS device-select, CLK bits) were looked up
// directly against a public hardware register reference (YAGCD
// chapter 5, "EXI - External Interface") for this addition.
//
// No actual EXI device protocol is modeled - there's no memory card,
// no IPL RTC/SRAM behind these registers. Writing TSTART completes
// immediately (the bit clears itself, like a real transfer finishing)
// leaving EXIDATA exactly as the caller wrote it, rather than a real
// attached device's response - the same "register accepted, real
// device behavior not modeled" simplification this project already
// uses for SI's command registers.
package exi

const (
	Base = 0xCC006800
	Size = 0x40

	numChannels   = 3
	channelStride = 0x14

	regCSR  = 0x00
	regData = 0x10

	bitTSTART = 1 << 0
)

type channel struct {
	csr  uint32
	data uint32
}

// EXI holds the External Interface's per-channel register state.
type EXI struct {
	channels [numChannels]channel
}

func New() *EXI { return &EXI{} }

func (e *EXI) Read32(offset uint32) uint32 {
	ch := int(offset / channelStride)
	if ch >= numChannels {
		return 0
	}
	c := &e.channels[ch]
	switch offset % channelStride {
	case regCSR:
		return c.csr
	case regData:
		return c.data
	default:
		return 0
	}
}

func (e *EXI) Write32(offset uint32, val uint32) {
	ch := int(offset / channelStride)
	if ch >= numChannels {
		return
	}
	c := &e.channels[ch]
	switch offset % channelStride {
	case regCSR:
		c.csr = val &^ bitTSTART // TSTART self-clears: no device to delay completion
	case regData:
		c.data = val
	}
}
