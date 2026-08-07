// Package atari2600 wires a from-scratch TIA and RIOT (see their own
// packages for why each is implemented the way it is) together with
// this project's existing NES 6502 core - the Atari 2600's 6507 CPU is
// the same 6502 core with fewer address pins bonded out, which the
// memory package's own address masking already reproduces, so no new
// CPU implementation is needed.
package atari2600

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/atari2600/memory"
	"github.com/Duc-inc/espaze/internal/systems/atari2600/riot"
	"github.com/Duc-inc/espaze/internal/systems/atari2600/tia"
	cpu "github.com/Duc-inc/espaze/internal/systems/nes/cpu"
)

const systemID = "atari2600"

const cpuClockHz = 1193182.0 // NTSC 2600 clock (TIA color clock / 3)

// Atari2600 wires the reused 6502 core to a from-scratch TIA and RIOT.
type Atari2600 struct {
	bus   *memory.Bus
	proc  *cpu.CPU
	video *tia.TIA
	riotC *riot.RIOT

	loaded bool
}

// New constructs a core with its stateless components ready; LoadROM
// still needs to be called before StepFrame does anything useful.
func New() core.Core {
	return &Atari2600{video: tia.New(), riotC: riot.New()}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the Atari 2600 core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "Atari 2600",
		Extensions:      []string{".a26", ".bin"},
		FramesPerSecond: cpuClockHz / (262 * 228 / 3),
		ScreenWidth:     tia.Width,
		ScreenHeight:    tia.Height,
	}
}

// Metadata implements core.Core.
func (a *Atari2600) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core. Atari 2600 cartridges are raw binary
// images - no header to parse.
func (a *Atari2600) LoadROM(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("atari2600: empty ROM")
	}

	a.video.Reset()
	a.riotC.Reset()
	a.bus = memory.New(data, a.video, a.riotC)
	a.proc = cpu.New(a.bus)
	a.loaded = true
	return nil
}

// Reset implements core.Core.
func (a *Atari2600) Reset() {
	if !a.loaded {
		return
	}
	a.video.Reset()
	a.riotC.Reset()
	a.proc.Reset()
}

// StepFrame implements core.Core: runs the CPU/TIA/RIOT together for
// exactly one 262-line raster, honoring WSYNC by advancing the TIA
// alone (matching real hardware halting the CPU, not the TIA's own
// free-running beam) until the next scanline starts.
func (a *Atari2600) StepFrame() error {
	if !a.loaded {
		return nil
	}

	for {
		if a.video.WSyncPending() {
			a.video.Step(1)
			a.riotC.Step(1)
		} else {
			cycles := a.proc.Step()
			a.video.Step(cycles)
			a.riotC.Step(cycles)
		}
		if a.video.FrameDone() {
			break
		}
	}

	return nil
}

// FrameBuffer implements core.Core.
func (a *Atari2600) FrameBuffer() *video.FrameBuffer { return a.video.FrameBuffer() }

// DrainAudio implements core.Core.
func (a *Atari2600) DrainAudio() ([]int16, int) {
	return a.video.DrainSamples(), 44100
}

// SetInput implements core.Core.
func (a *Atari2600) SetInput(state emuinput.State) { a.bus.SetButtons(state) }
