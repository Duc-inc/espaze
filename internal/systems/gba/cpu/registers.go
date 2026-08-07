package cpu

// CPSR mode bits (bits 0-4) - the 5 modes this project implements
// registers for; System/Undefined modes aren't given their own banked
// registers since no code this project targets relies on them.
const (
	modeUser       = 0x10
	modeFIQ        = 0x11
	modeIRQ        = 0x12
	modeSupervisor = 0x13
	modeAbort      = 0x17
)

// CPSR flag/control bits.
const (
	FlagN     uint32 = 1 << 31 // negative
	FlagZ     uint32 = 1 << 30 // zero
	FlagC     uint32 = 1 << 29 // carry
	FlagV     uint32 = 1 << 28 // overflow
	FlagIRQD  uint32 = 1 << 7  // IRQ disable
	FlagFIQD  uint32 = 1 << 6  // FIQ disable
	FlagThumb uint32 = 1 << 5  // Thumb state
)

// registers holds the ARM7TDMI's visible register file: 16 general
// registers (R13=SP, R14=LR, R15=PC by convention) plus banked copies
// for FIQ/IRQ/Supervisor/Abort modes (each mode has its own R13/R14,
// and FIQ additionally banks R8-R12 - a detail no game this project
// targets is known to depend on, so those 5 aren't banked here).
type registers struct {
	R    [16]uint32
	CPSR uint32

	bankedSP, bankedLR [5]uint32 // indexed by modeIndex(cpsr mode)
	spsr               [5]uint32
}

func modeIndex(mode uint32) int {
	switch mode {
	case modeFIQ:
		return 0
	case modeIRQ:
		return 1
	case modeSupervisor:
		return 2
	case modeAbort:
		return 3
	default:
		return 4 // User/System share one slot that's never actually read
	}
}

func (r *registers) mode() uint32 { return r.CPSR & 0x1F }

func (r *registers) getFlag(flag uint32) bool { return r.CPSR&flag != 0 }

func (r *registers) setFlag(flag uint32, on bool) {
	if on {
		r.CPSR |= flag
	} else {
		r.CPSR &^= flag
	}
}

func (r *registers) thumb() bool { return r.getFlag(FlagThumb) }

// switchMode banks out the current mode's SP/LR and banks in the
// target mode's, then updates the CPSR mode bits.
func (r *registers) switchMode(target uint32) {
	cur := modeIndex(r.mode())
	r.bankedSP[cur] = r.R[13]
	r.bankedLR[cur] = r.R[14]

	next := modeIndex(target)
	r.R[13] = r.bankedSP[next]
	r.R[14] = r.bankedLR[next]

	r.CPSR = r.CPSR&^0x1F | target
}
