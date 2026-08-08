// Package ipl provides this project's own stand-in for the GameCube's
// IPL (boot ROM), which this project doesn't have (Nintendo's is
// copyrighted firmware, not something a from-scratch project can
// include). Real IPL firmware reads the disc, loads the DOL, sets up
// the MMU/BAT registers and a handful of OS library entry points, then
// jumps to the game's entry point. This package does the two pieces
// that are actually tractable without real firmware - parsing and
// loading the DOL (internal/systems/gamecube/disc) and setting the
// CPU's initial register state - plus a syscall trap mechanism
// (powerpc.CPU.SyscallHandler) for future high-level emulation of the
// small set of OS calls real games actually make, rather than
// interpreting IPL code this project will never have.
package ipl

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/gamecube/disc"
	"github.com/Duc-inc/espaze/internal/systems/powerpc"
)

// stackTop is where this project starts the initial stack pointer -
// near the top of MEM1's 24MB, leaving room below it for the DOL's
// own sections (real IPL computes this dynamically based on the
// loaded executable; this project fixes it, which is fine for the
// common case of a DOL that doesn't reach that high itself).
const stackTop = 0x817FFFF0

// Boot parses a disc image's header and DOL, loads every DOL section
// into memory, and sets the CPU up to start executing at the game's
// entry point with an initialized stack pointer (r1, PowerPC's own
// stack-pointer convention).
func Boot(image []byte, mem disc.Writer, cpu *powerpc.CPU) (disc.Header, error) {
	header, err := disc.ParseHeader(image)
	if err != nil {
		return disc.Header{}, err
	}
	if int(header.DOLOffset) >= len(image) {
		return disc.Header{}, fmt.Errorf("ipl: DOL offset %#08x is past the end of the image", header.DOLOffset)
	}

	dol := disc.ParseDOL(image[header.DOLOffset:])
	// Section file offsets inside a DOL are relative to the DOL's own
	// start, not the disc image's start, so re-anchor them before loading.
	for i := range dol.Sections {
		dol.Sections[i].FileOffset += header.DOLOffset
	}
	dol.LoadInto(image, mem)

	cpu.SetGPR(1, stackTop)
	cpu.SetPC(dol.Entry)
	return header, nil
}
