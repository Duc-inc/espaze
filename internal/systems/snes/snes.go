// Package snes wires a from-scratch 65816 main CPU, PPU, and a
// simplified SPC700+DSP audio coprocessor together into a core.Core
// implementation - the Super Nintendo. See each subpackage's own doc
// comment for what's simplified and why; the spc700 package carries
// the same "own internally consistent encoding" caveat as this
// project's Neo Geo Pocket TLCS900H core, since SPC700's exact
// instruction encoding is far less documented than the 65816's.
package snes

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/emulation/core"
	emuinput "github.com/Duc-inc/espaze/internal/emulation/input"
	"github.com/Duc-inc/espaze/internal/emulation/video"

	"github.com/Duc-inc/espaze/internal/systems/snes/audio"
	"github.com/Duc-inc/espaze/internal/systems/snes/cpu"
	"github.com/Duc-inc/espaze/internal/systems/snes/dsp"
	"github.com/Duc-inc/espaze/internal/systems/snes/memory"
	"github.com/Duc-inc/espaze/internal/systems/snes/ppu"
	"github.com/Duc-inc/espaze/internal/systems/snes/spc700"
)

const systemID = "snes"

const cpuClockHz = 21477272.0
const cyclesPerFrame = 262 * 1364 // matches ppu's own scanline timing assumption

const spcClockHz = 1024000.0

// SNES wires the 65816/PPU/SPC700/DSP together.
type SNES struct {
	bus   *memory.Bus
	proc  *cpu.CPU
	video *ppu.PPU

	sound   *dsp.DSP
	spc     *spc700.CPU
	spcBus  *audio.Bus
	ports   audio.Ports
	spcLeft float64

	loaded bool
}

// New constructs a core with its stateless components ready; LoadROM
// still needs to be called before StepFrame does anything useful.
func New() core.Core {
	return &SNES{video: ppu.New(), sound: dsp.New()}
}

func init() {
	core.Register(Metadata(), New)
}

// Metadata describes the SNES core for the registry and scanner.
func Metadata() core.Metadata {
	return core.Metadata{
		ID:              systemID,
		Name:            "Super Nintendo",
		Extensions:      []string{".sfc", ".smc"},
		FramesPerSecond: cpuClockHz / cyclesPerFrame,
		ScreenWidth:     ppu.Width,
		ScreenHeight:    ppu.Height,
	}
}

// Metadata implements core.Core.
func (s *SNES) Metadata() core.Metadata { return Metadata() }

// LoadROM implements core.Core. SNES cartridges are raw binary images
// - no header this project parses (some real dumps have a 512-byte
// copier header; that isn't stripped automatically here).
func (s *SNES) LoadROM(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("snes: empty ROM")
	}

	s.video.Reset()
	s.sound.Reset()
	s.ports = audio.Ports{}

	s.spcBus = audio.New(s.sound, &s.ports)
	s.spc = spc700.New(s.spcBus)

	s.bus = memory.New(data, s.video, &s.ports)
	s.proc = cpu.New(s.bus)

	s.spcLeft = 0
	s.loaded = true
	return nil
}

// Reset implements core.Core.
func (s *SNES) Reset() {
	if !s.loaded {
		return
	}
	s.video.Reset()
	s.sound.Reset()
	s.spcBus.Reset()
	s.spc.Reset()
	s.proc.Reset()
	s.spcLeft = 0
}

// StepFrame implements core.Core: runs the 65816/PPU/SPC700 together
// for exactly one 262-line raster.
func (s *SNES) StepFrame() error {
	if !s.loaded {
		return nil
	}

	spent := 0
	for spent < cyclesPerFrame {
		cycles := s.proc.Step()

		if irq := s.video.Step(cycles); irq&ppu.IRQVBlank != 0 {
			s.proc.TriggerNMI()
		}
		s.stepSPC(cycles)

		spent += cycles
	}

	return nil
}

// stepSPC advances the audio coprocessor by its own clock's share of
// the main CPU cycles just spent.
func (s *SNES) stepSPC(mainCycles int) {
	s.spcLeft += float64(mainCycles) * (spcClockHz / cpuClockHz)
	for s.spcLeft > 0 {
		cycles := s.spc.Step()
		s.sound.Step(cycles)
		s.spcLeft -= float64(cycles)
	}
}

// FrameBuffer implements core.Core.
func (s *SNES) FrameBuffer() *video.FrameBuffer { return s.video.FrameBuffer() }

// DrainAudio implements core.Core.
func (s *SNES) DrainAudio() ([]int16, int) {
	return s.sound.DrainSamples(), dsp.SampleRate
}

// SetInput implements core.Core.
func (s *SNES) SetInput(state emuinput.State) { s.bus.SetButtons(state) }
