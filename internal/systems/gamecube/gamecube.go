// Package gamecube wires the from-scratch PowerPC CPU
// (internal/systems/powerpc) to this platform's own physical memory
// map (internal/systems/gamecube/memory). It deliberately does NOT
// implement core.Core and is NOT registered with the emulation
// registry - see the memory package's own doc comment for why: there
// is no GPU pipeline and no disc reader yet, so this can run
// arbitrary PowerPC code against RAM but cannot boot or display
// anything from a real game. This is groundwork for a future GameCube
// core, not a shipped feature.
package gamecube

import (
	"github.com/Duc-inc/espaze/internal/systems/gamecube/memory"
	"github.com/Duc-inc/espaze/internal/systems/powerpc"
)

// GameCube wires the CPU to its memory bus.
type GameCube struct {
	bus  *memory.Bus
	proc *powerpc.CPU
}

// New wires a fresh CPU and memory bus together and resets both.
func New() *GameCube {
	g := &GameCube{bus: memory.New()}
	g.proc = powerpc.New(g.bus)
	return g
}

// Reset clears RAM and every register.
func (g *GameCube) Reset() {
	g.bus.Reset()
	g.proc.Reset()
}

// Step executes exactly one PowerPC instruction, returning an
// approximate cycle cost.
func (g *GameCube) Step() int { return g.proc.Step() }

// LoadAt copies data directly into MEM1 at the given physical address
// - a stand-in for the disc-loading process real hardware's IPL (boot
// ROM) performs, since this project has no disc reader yet.
func (g *GameCube) LoadAt(addr uint32, data []byte) {
	for i, v := range data {
		g.bus.Write8(addr+uint32(i), v)
	}
}
