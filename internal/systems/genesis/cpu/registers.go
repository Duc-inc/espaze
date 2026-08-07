package cpu

// Status register (CCR) flag bits.
const (
	FlagC uint16 = 1 << 0 // carry
	FlagV uint16 = 1 << 1 // overflow
	FlagZ uint16 = 1 << 2 // zero
	FlagN uint16 = 1 << 3 // negative
	FlagX uint16 = 1 << 4 // extend (a "sticky" copy of carry some multi-word arithmetic ops use)
)

// System byte bits (the upper 8 bits of SR, privileged to change).
const (
	srSupervisor uint16 = 1 << 13
	srIPMask     uint16 = 0x0700 // interrupt priority mask, bits 8-10
)

// registers holds the 68000's full visible state: 8 data registers, 8
// address registers (A7 is the active stack pointer - which physical
// register that means depends on supervisor mode, see usp/ssp below),
// the program counter, and the status register.
type registers struct {
	D [8]uint32
	A [8]uint32

	PC uint32
	SR uint16

	usp, ssp uint32 // the *other* stack pointer, not currently aliased into A[7]
}

func (r *registers) supervisor() bool { return r.SR&srSupervisor != 0 }

func (r *registers) setFlag(flag uint16, on bool) {
	if on {
		r.SR |= flag
	} else {
		r.SR &^= flag
	}
}

func (r *registers) getFlag(flag uint16) bool { return r.SR&flag != 0 }

// setNZ sets Negative/Zero from a just-computed result - the flag
// pattern shared by nearly every data-processing instruction, sized to
// only look at the bits the operation actually worked on.
func (r *registers) setNZ32(v uint32) {
	r.setFlag(FlagN, int32(v) < 0)
	r.setFlag(FlagZ, v == 0)
}

// interruptMask returns the current interrupt priority mask (0-7);
// only interrupts with a strictly higher priority get through.
func (r *registers) interruptMask() byte {
	return byte((r.SR & srIPMask) >> 8)
}

// enterSupervisor switches the currently-active stack pointer (A[7])
// from USP to SSP if not already in supervisor mode - real hardware
// does this automatically on any exception/interrupt entry.
func (r *registers) enterSupervisor() {
	if r.supervisor() {
		return
	}
	r.usp = r.A[7]
	r.A[7] = r.ssp
	r.SR |= srSupervisor
}
