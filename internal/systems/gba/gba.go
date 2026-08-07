// Package gba wires a from-scratch ARM7TDMI, PPU, and Direct Sound APU
// together into a core.Core implementation - the Game Boy Advance. See
// each subpackage's own doc comment for what's simplified and why.
package gba

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/gba/apu"
	"github.com/Duc-inc/espaze/internal/systems/gba/cpu"
	"github.com/Duc-inc/espaze/internal/systems/gba/memory"
	"github.com/Duc-inc/espaze/internal/systems/gba/ppu"
)

const systemID = "gba"

const cpuClockHz = 16777216.0
const cyclesPerFrame = 228 * 308 // matches ppu's own scanline timing assumption

// GBA wires the CPU/PPU/APU/bus together.
type GBA struct {
	bus   *memory.Bus
	proc  *cpu.CPU
	video *ppu.PPU
	sound *apu.APU

	loaded bool
}

// New constructs a core with its stateless components ready; LoadROM
// still needs to be called before StepFrame does anything useful.
func New() core.Core {
	return &GBA{video: ppu.New(), sound: apu.New()}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the GBA core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "Game Boy Advance",
		Extensions:      []string{".gba"},
		FramesPerSecond: cpuClockHz / cyclesPerFrame,
		ScreenWidth:     ppu.Width,
		ScreenHeight:    ppu.Height,
	}
}

// Metadata implements core.Core.
func (g *GBA) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core. GBA cartridges are raw binary images -
// no header to parse.
func (g *GBA) LoadROM(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("gba: empty ROM")
	}

	g.video.Reset()
	g.sound.Reset()
	g.bus = memory.New(data, g.video, g.sound)
	g.proc = cpu.New(g.bus)
	g.loaded = true
	return nil
}

// Reset implements core.Core.
func (g *GBA) Reset() {
	if !g.loaded {
		return
	}
	g.video.Reset()
	g.sound.Reset()
	g.proc.Reset()
}

// StepFrame implements core.Core: runs the CPU/PPU/APU/timers together
// for exactly one 228-line raster.
func (g *GBA) StepFrame() error {
	if !g.loaded {
		return nil
	}

	spent := 0
	for spent < cyclesPerFrame {
		cycles := g.proc.Step()

		if irq := g.video.Step(cycles); irq&ppu.IRQVBlank != 0 {
			g.bus.RaiseVBlank()
		}
		g.bus.StepTimers(cycles)
		g.sound.Step(cycles)

		if g.bus.InterruptPending() {
			g.proc.TriggerIRQ()
		}

		spent += cycles
	}

	return nil
}

// FrameBuffer implements core.Core.
func (g *GBA) FrameBuffer() *video.FrameBuffer { return g.video.FrameBuffer() }

// DrainAudio implements core.Core.
func (g *GBA) DrainAudio() ([]int16, int) {
	return g.sound.DrainSamples(), apu.SampleRate
}

// SetInput implements core.Core.
func (g *GBA) SetInput(state emuinput.State) { g.bus.SetButtons(state) }
