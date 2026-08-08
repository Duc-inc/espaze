// Package pi implements the GameCube's Processor Interface: the real
// register block (YAGCD chapter 5, section 5.4) that ORs every
// peripheral's interrupt cause into the single external interrupt line
// the PowerPC CPU actually has (see powerpc's exceptions.go). Base
// address and the INTSR/INTMR bit-to-peripheral table below were
// looked up directly against that source for this addition; every
// other GameCube peripheral package in this project reports its own
// interrupt causes into PI via SetCause instead of raising the CPU
// exception directly, matching how real hardware layers this.
package pi

const (
	Base = 0xCC003000
	Size = 0x08

	regINTSR = 0x00 // interrupt cause (read-only status, see doc below)
	regINTMR = 0x04 // interrupt mask
)

// Interrupt cause bits, INTSR/INTMR's shared bit assignments.
const (
	BitError    = 0
	BitRSW      = 1
	BitDI       = 2
	BitSI       = 3
	BitEXI      = 4
	BitAI       = 5
	BitDSP      = 6
	BitMEM      = 7
	BitVI       = 8
	BitPEToken  = 9
	BitPEFinish = 10
	BitCP       = 11
	BitDebug    = 12
	BitHSP      = 13
)

// PI holds the Processor Interface's cause/mask state.
type PI struct {
	cause uint32
	mask  uint32
}

func New() *PI { return &PI{} }

// SetCause sets or clears one interrupt-cause bit - callers are every
// other peripheral package (vi.VI.AnyInterruptActive, ai.AI.Interrupting,
// ...) reporting their own current interrupt state each GameCube.Step.
// Real hardware's INTSR bits reflect the source peripheral's own status
// register (e.g. VI's DI0-3 active bits) rather than being writable
// through PI directly, which is why this project doesn't expose a
// generic "clear via INTSR write" path (Write32 below is INTMR-only).
func (p *PI) SetCause(bit uint, active bool) {
	if active {
		p.cause |= 1 << bit
	} else {
		p.cause &^= 1 << bit
	}
}

// Pending reports whether any unmasked cause is active - what should
// gate CPU.RaiseExternalInterrupt (see gamecube.go's Step).
func (p *PI) Pending() bool { return p.cause&p.mask != 0 }

func (p *PI) Read32(offset uint32) uint32 {
	switch offset {
	case regINTSR:
		return p.cause
	case regINTMR:
		return p.mask
	default:
		return 0
	}
}

func (p *PI) Write32(offset uint32, val uint32) {
	if offset == regINTMR {
		p.mask = val
	}
	// regINTSR: read-only status here, see SetCause's doc comment.
}
