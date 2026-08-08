// Package gamecube wires the from-scratch PowerPC CPU
// (internal/systems/powerpc) to this platform's own physical memory
// map (internal/systems/gamecube/memory) plus the real VI/SI/DI/AI
// hardware register peripherals. It deliberately does NOT implement
// core.Core and is NOT registered with the emulation registry - see
// the memory package's own doc comment for why: there is still no GX
// pipeline wired in here, so this can run arbitrary PowerPC code and
// drive real peripheral registers but cannot display a real game yet.
// This is groundwork for a future GameCube core, not a shipped
// feature.
package gamecube

import (
	"github.com/Duc-inc/espaze/internal/systems/gamecube/ai"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/audio"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/di"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/memory"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/si"
	"github.com/Duc-inc/espaze/internal/systems/gamecube/vi"
	"github.com/Duc-inc/espaze/internal/systems/powerpc"
)

// GameCube wires the CPU to its memory bus and hardware peripherals.
type GameCube struct {
	bus  *memory.Bus
	proc *powerpc.CPU

	VI *vi.VI
	SI *si.SI
	DI *di.DI
	AI *ai.AI

	// Audio is the sample mixer AI's PSTAT bit gates (Step below) - AI
	// itself only tracks streaming on/off and volume, it has no sample
	// data of its own; a caller feeds real channel data into Audio via
	// SetChannel/SetADPCMChannel same as before this field existed.
	// AI.Volume()'s L/R scaling isn't applied to Audio's output - Mixer
	// only produces one mixed mono sample, not a stereo pair.
	Audio *audio.Mixer
}

// New wires a fresh CPU, memory bus, and VI/SI/DI/AI peripherals
// together and resets the CPU/RAM. discImage may be nil if there's no
// disc to serve DI reads from yet.
func New(discImage []byte) *GameCube {
	g := &GameCube{bus: memory.New()}
	g.proc = powerpc.New(g.bus)

	g.VI = vi.New()
	g.SI = si.New()
	g.DI = di.New(discImage, g.bus)
	g.AI = ai.New()
	g.Audio = audio.New()

	g.bus.Attach(vi.Base, vi.Size, g.VI)
	g.bus.Attach(si.Base, si.Size, g.SI)
	g.bus.Attach(di.Base, di.Size, g.DI)
	g.bus.Attach(ai.Base, ai.Size, g.AI)

	return g
}

// Reset clears RAM and every register. Peripheral state (SI channel
// data, AI volume, etc.) is left as-is - callers that want that reset
// too should reconstruct via New.
func (g *GameCube) Reset() {
	g.bus.Reset()
	g.proc.Reset()
}

// Step executes exactly one PowerPC instruction, then ticks VI/AI and
// delivers a real external interrupt exception (see powerpc's
// exceptions.go) if either just fired. Real hardware ticks VI/AI on
// their own pixel/sample clocks, independent of CPU instruction
// count; tying them to one Step call each is this project's own
// simplification, not a timing claim.
func (g *GameCube) Step() int {
	cycles := g.proc.Step()
	if g.VI.Step() {
		g.proc.RaiseExternalInterrupt()
	}
	if g.AI.Step() {
		g.proc.RaiseExternalInterrupt()
	}
	if g.AI.Playing() {
		g.Audio.Step(1)
	}
	return cycles
}

// LoadAt copies data directly into MEM1 at the given physical address
// - a stand-in for the disc-loading process real hardware's IPL (boot
// ROM) performs.
func (g *GameCube) LoadAt(addr uint32, data []byte) {
	for i, v := range data {
		g.bus.Write8(addr+uint32(i), v)
	}
}
